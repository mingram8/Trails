package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/mingram/trail/kml"
	"github.com/mingram/trail/osm"
	"github.com/umahmood/haversine"
	"io/ioutil"
	"strings"
	"sync"

	//"github.com/AvraamMavridis/randomcolor"

	"log"
	"os"
)

type Geometry struct {
	Tipo        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}
type Properties struct {
	Stroke        string  `json:"stroke"`
	StrokeWidth   float64 `json:"stroke-width"`
	StrokeOpacity float64 `json:"stroke-opacity"`
	Fill          string  `json:"fill"`
	Name          string  `json:"name"`
	FillOpacity   float64 `json:"fill-opacity"`
}
type Feature struct {
	Tipo       string     `json:"type"`
	Geometry   Geometry   `json:"geometry"`
	Properties Properties `json:"properties"`
}

type GeoJson struct {
	Tipo     string    `json:"type"`
	Features []Feature `json:"features"`
}

func main() {
	osmFile := flag.String("file", "frederick-county.osm", "osm file")
	activity := flag.String("activity", "any", "Type of activity")
	fileType := flag.String("type", "kml", "Type of activity")

	flag.Parse()

	xmlFile, err := os.Open(*osmFile)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened " + *osmFile)
	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	// we initialize our Users array
	var osm openStreetMap.Osm
	// we unmarshal our byteArray which contains our
	// xmlFiles content into 'users' which we defined above
	xml.Unmarshal(byteValue, &osm)

	KML := kml.NewKml(*activity+" Trails", "Trails")
	KML.AddStyle("00FFFF", "FF00FFFF", 4)
	KML.AddStyle("FFD700", "FFFFD700", 4)
	KML.AddStyle("3333ff", "FF3333ff", 4)
	KML.AddStyle("d699ff", "FFd699ff", 4)
	KML.AddStyle("00FFFF", "FF00FFFF", 4)
	KML.AddStyle("b3b3ff", "FFb3b3ff", 4)
	KML.AddStyle("FF4DFF", "FFFF4DFF", 4)
	KML.AddStyle("ff99cc", "FFff99cc", 4)
	KML.AddStyle("ffff66", "FFffff66", 4)

	var mtnBikes []openStreetMap.Way
	var nodes [][]openStreetMap.Node
	var ski openStreetMap.Ski
	var mtnbike openStreetMap.Mtnbike
	var foot openStreetMap.Foot
	var wg sync.WaitGroup

	for _, way := range osm.Ways {
		types := make(map[string]string)
		add := false
		for _, tag := range way.Tags {
			key, value := tag.Key, tag.Value
			types[key] = value
		}
		way.Name = types["name"]
		if types["ski"] == "yes" || types["piste:type"] == "downhill" {
			ski = openStreetMap.Ski{types["piste:difficulty"], "allowed", types["piste:type"]}
			if strings.Index(*activity, "ski") != -1 || *activity == "any" {
				way.Ski = ski
				add = true
			}
		}
		if types["bicycle"] == "yes" || types["bicycle"] == "designated" && types["highway"] == "path" {
			if types["highway"] == "path" {
				diff := types["mtb:scale:imba"]
				description := types["description"]
				surface := types["surface"]
				if surface == "" {
					surface = "unknown"
				}
				if description == "" {
					description = "allowed"
				}
				if diff == "" {
					diff = types["mtb:scale"]
				}
				mtnbike = openStreetMap.Mtnbike{diff, description, surface}
				if strings.Index(*activity, "bike") != -1 || *activity == "any" {
					way.Mtnbike = mtnbike
					add = true
				}
			}
		}
		if types["foot"] == "yes" || types["foot"] == "designated" || types["foot"] == "permissive" {
			if types["highway"] == "path" {
				foot = openStreetMap.Foot{"none", "foot", types["surface"]}
				if strings.Index(*activity, "hike") != -1 || *activity == "any" {
					way.Foot = foot
					add = true
				}
			}
		} else if types["highway"] == "path" {
			foot = openStreetMap.Foot{"none", "unknown", types["surface"]}
			if strings.Index(*activity, "walk") != -1 || *activity == "any" {
				way.Foot = foot
				add = true
			}
		}
		if add == true {
			mtnBikes = append(mtnBikes, way)
		}

	}
	log.Print("Number of trails: " + fmt.Sprintf("%v", len(mtnBikes)))

	for _, mtnBike := range mtnBikes {
		var no []openStreetMap.Node
		wg.Add(len(mtnBike.Nds))
		for _, nd := range mtnBike.Nds {
			messages := make(chan openStreetMap.Node)
			go matchNode(nd, mtnBike, osm, messages, &wg)
			newNode := <-messages
			no = append(no, newNode)

		}
		nodes = append(nodes, no)
	}
	wg.Wait()
	for _, mtnBike := range mtnBikes {
		for _, mtnBike2 := range mtnBikes {
			_, canBe := openStreetMap.CombineWays(mtnBike, mtnBike2)
			if canBe == true {
				//log.Print(mtnBike, mtnBike2)
			}
		}
	}
	var features []Feature
	geojson := GeoJson{}
	geojson.Tipo = "FeatureCollection"

	for _, node := range nodes {
		if *fileType == "geojson" {
			color, _ := GetColor(node[0])
			var featuresLocal []Feature

			geojsonLocal := GeoJson{}
			geojsonLocal.Tipo = "FeatureCollection"
			var coordinates [][]float64
			for _, nd := range node {
				coordinates = append(coordinates, []float64{nd.Lon, nd.Lat})

			}

			geometry := Geometry{"LineString", coordinates}
			feature := Feature{}
			feature.Tipo = "Feature"
			feature.Properties = Properties{Name: node[0].Name, Stroke: color, Fill: "#FFF", FillOpacity: .5, StrokeOpacity: 1.0, StrokeWidth: 2}
			feature.Geometry = geometry
			featuresLocal = append(featuresLocal, feature)
			features = append(features, feature)
			geojsonLocal.Features = featuresLocal
			start, end := []string{fmt.Sprintf("%f", node[0].Lon), fmt.Sprintf("%f", node[0].Lat)}, []string{fmt.Sprintf("%f", node[len(node)-1].Lon), fmt.Sprintf("%f", node[len(node)-1].Lat)} // s == "123.456000"

			name := strings.Replace(node[0].Name, "/", "-", -1)
			os.MkdirAll("geojson/trails", os.ModePerm)
			json, _ := json.Marshal(geojsonLocal)
			err = ioutil.WriteFile("geojson/trails/"+name+"-Start-"+start[0]+","+start[1]+"End-"+end[0]+","+end[1]+".json", json, 0644)

		} else if *fileType == "kml" {
			KMLlocal := kml.NewKml(node[0].Name, "Trails")
			KMLlocal.AddStyle("00FFFF", "FF00FFFF", 4)
			KMLlocal.AddStyle("FFD700", "FFFFD700", 4)
			KMLlocal.AddStyle("3333ff", "FF3333ff", 4)
			KMLlocal.AddStyle("d699ff", "FFd699ff", 4)
			KMLlocal.AddStyle("00FFFF", "FF00FFFF", 4)
			KMLlocal.AddStyle("b3b3ff", "FFb3b3ff", 4)
			KMLlocal.AddStyle("FF4DFF", "FFFF4DFF", 4)
			KMLlocal.AddStyle("ff99cc", "FFff99cc", 4)
			KMLlocal.AddStyle("ffff66", "FFffff66", 4)
			var kmlCoordinates [][]float64
			color, tipo := GetColor(node[0])
			var totalDistance float64
			var oldNode openStreetMap.Node
			nahs := []openStreetMap.Node{}
			for _, nd := range node {
				nahs = append(nahs, nd)
				kmlCoordinates = append(kmlCoordinates, []float64{nd.Lon, nd.Lat, 0.0})
				if oldNode != (openStreetMap.Node{}) {
					d := distance(nd.Lon, nd.Lat, oldNode.Lon, oldNode.Lat)
					totalDistance += d
				}
				oldNode = nd

			}
			os.MkdirAll("kmls/trails", os.ModePerm)
			start, end := []string{fmt.Sprintf("%f", node[0].Lon), fmt.Sprintf("%f", node[0].Lat)}, []string{fmt.Sprintf("%f", node[len(node)-1].Lon), fmt.Sprintf("%f", node[len(node)-1].Lat)} // s == "123.456000"

			name := strings.Replace(node[0].Name, "/", "-", -1)
			description := "Type: " + tipo + "\n" +
				"Total Distance: " + fmt.Sprintf("%f", totalDistance) + " km"

			KMLlocal.AddPlacemark(name, color, description, kmlCoordinates, nahs, "false")
			KML.AddPlacemark(name, color, description, kmlCoordinates, nahs, "true")
			KMLlocal.SaveFile("kmls/trails/" + name + "-Start-" + start[0] + "," + start[1] + "End-" + end[0] + "," + end[1] + ".kml")
		}
	}

	//j :=  []byte(json)
	//log.Print(string(j))

	if *fileType == "kml" {
		//KML.SortPlacemarkers()
		//log.Print(len(KML.Placemarks))

		KML.SaveFile("kmls/ALL_TRAILS.kml")
	} else if *fileType == "geojson" {
		geojson.Features = features
		json, _ := json.Marshal(geojson)
		err = ioutil.WriteFile("geojson/ALL_TRAILS.json", json, 0644)
	}

}

func matchNode(nd openStreetMap.Nd, way openStreetMap.Way, osm openStreetMap.Osm, c chan openStreetMap.Node, wg *sync.WaitGroup) {

	for _, node := range osm.Nodes {
		if nd.Ref == node.Id {
			node.Name = way.Name
			node.Wayid = way.Id
			node.Ski = way.Ski
			node.Mtnbike = way.Mtnbike
			node.Foot = way.Foot
			c <- node
			wg.Done()
			return
		}
	}
}

func sortTag(tag openStreetMap.Tag) (string, string) {
	return tag.Key, tag.Value
}
func GetColor(tipo openStreetMap.Node) (string, string) {
	var bike, ski, foot bool
	if tipo.Mtnbike != (openStreetMap.Mtnbike{}) {
		bike = true
	}
	if tipo.Ski != (openStreetMap.Ski{}) {
		ski = true
	}
	if tipo.Foot.Tipo == "foot" {
		foot = true
	}
	if bike == true && ski == true && foot == true {
		return "#00FFFF", "Ski, Bike, Hike"
	} else if bike == true && foot == true {
		return "#FFD700", "Bike, Hike"
	} else if ski == true && foot == true {
		return "#3333ff", "Ski, Hike"
	} else if ski == true && bike == true {
		return "#d699ff", "Ski, Bike"
	} else if ski == true {
		return "#b3b3ff", "Ski"
	} else if bike == true {
		return "#FF4DFF", "Bike"
	} else if foot == true {
		return "#ff99cc", "Hike"
	} else {
		return "#ffff66", "Unknown"
	}

}
func distance(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64 {
	pos1 := haversine.Coord{Lat: lat1, Lon: lon1} //
	pos2 := haversine.Coord{Lat: lat2, Lon: lon2} //
	_, km := haversine.Distance(pos1, pos2)

	return km
}
