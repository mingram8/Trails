package kml

import (
	"crypto/rand"
	"encoding/xml"
	"fmt"
	"github.com/mingram/trail/osm"
	"github.com/umahmood/haversine"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

type Timespan struct {
	XMLName xml.Name `xml:"Timespan"`
	Begin   string   `xml:"begin"`
	End     string   `xml:"end"`
}
type Linestring struct {
	XMLName      xml.Name    `xml:"LineString"`
	Coords       string      `xml:"coordinates"`
	Extrude      int         `xml:"extrude"`
	AltitudeMode string      `xml:"altitudeMode"`
	Tessellate   int         `xml:"tessellate"`
	Coordinates  [][]float64 `"json:"coordinates`
}

type Linestyle struct {
	XMLName xml.Name `xml:"LineStyle"`
	Color   string   `xml:"color"`
	Width   int      `xml:"width"`
}
type Style struct {
	XMLName   xml.Name  `xml:"Style"`
	Id        string    `xml:"id,attr"`
	Linestyle Linestyle `xml:"LineStyle"`
}
type Placemark struct {
	XMLName     xml.Name             `xml:"Placemark"`
	Name        string               `xml:"name"`
	Id          string               `xml:"id"`
	StyleUrl    string               `xml:"styleUrl"`
	Description string               `xml:"description"`
	Linestring  Linestring           `xml:"LineString"`
	Nodes       []openStreetMap.Node `"json:"node`
}
type Kml struct {
	XMLName     xml.Name    `xml:"Document"`
	Name        string      `xml:"name"`
	Description string      `xml:"description"`
	Timespan    Timespan    `xml:"Timespan"`
	Style       []Style     `xml:"Style"`
	Placemarks  []Placemark `xml:"Placemark"`
}
type File struct {
	XMLName xml.Name `xml:"kml"`
	Kml     Kml      `xml:"Document"`
}

func NewKml(name string, description string) Kml {
	var kml Kml

	defaultLinestyle := Linestyle{Color: "ff000000", Width: 4}
	defaultStyle := Style{Id: "default", Linestyle: defaultLinestyle}
	kml.Style = append(kml.Style, defaultStyle)
	kml.Name = name
	kml.Description = description

	return kml
}
func ReadKML(reader []byte) Kml {
	var kml File
	// we unmarshal our byteArray which contains our
	// xmlFiles content into 'users' which we defined above
	xml.Unmarshal(reader, &kml)

	for i, placemark := range kml.Kml.Placemarks {
		var coordinates [][]float64
		coords := placemark.Linestring.Coords
		c := strings.Split(coords, "\n")
		for _, co := range c {
			coms := strings.Split(co, ",")
			var row []float64
			if len(coms) > 1 {
				for _, num := range coms {
					f, _ := strconv.ParseFloat(strings.Trim(num, " "), 64)
					log.Print(f)
					row = append(row, f)
				}
				coordinates = append(coordinates, row)
			}
		}
		kml.Kml.Placemarks[i].Linestring.Coordinates = coordinates
	}
	return kml.Kml
}

//func (kml *Kml) SetLineColor(color string, index int) {
//	kml.Style.Linestyle.Color = color
//}
func (kml *Kml) SetName(name string, index int) {
	kml.Placemarks[index].Name = name
}
func (kml *Kml) AddStyle(id string, color string, width int) {
	defaultLinestyle := Linestyle{Color: color, Width: width}
	defaultStyle := Style{Id: id, Linestyle: defaultLinestyle}
	kml.Style = append(kml.Style, defaultStyle)
}
func (kml *Kml) AddPlacemark(name string, styleUrl string, description string, coords [][]float64, nodes []openStreetMap.Node, str string) {
	var num int
	for _, mark := range kml.Placemarks {
		if name == mark.Name {
			num++
		}
	}
	var placemark Placemark
	placemark.Name = name
	placemark.StyleUrl = styleUrl
	placemark.Description = description
	var linestring Linestring
	linestring.Coordinates = [][]float64{[]float64{0.0, 0.0}, []float64{0.0, 0.0}}
	placemark.Linestring = linestring

	if name == "Catoctin National Recreation Trail" && num == 0 && str == "true" {
		log.Print(coords[0][1], coords[0][0])
		log.Print(coords[len(coords)-1][1], coords[len(coords)-1][0])

	}

	if name == "Catoctin National Recreation Trail" && num == 2 && str == "true" {
		log.Print(coords[0][1], coords[0][0])
		log.Print(coords[len(coords)-1][1], coords[len(coords)-1][0])

	}

	if strings.Index(name, "Blue Balls") != -1 || strings.Index(name, "Access to Freeride") != -1 {

		if styleUrl == "" {
			styleUrl = "default"
		}
		if description == "" {
			description = "default placemark"
		}
		if name == "" {
			name = "default placemark"
		}
		//placemark.Nodes = nodes

		placemark.Id = fmt.Sprintf("%v", num)
		linestring.Coordinates = coords
		linestring.AltitudeMode = "clampToGround"
		linestring.Tessellate = 1
		linestring.Extrude = 1
		placemark.Linestring = linestring

		kml.Placemarks = append(kml.Placemarks, placemark)
	}
}

func (kml *Kml) ConvertCoords() {
	for i, placemark := range kml.Placemarks {
		var coordinates string
		for _, coord := range placemark.Linestring.Coordinates {
			for i, c := range coord {
				s := fmt.Sprintf("%f", c) // s == "123.456000"
				remainder := i % 3
				if remainder != 2 {
					coordinates += s + ","
				} else {
					coordinates += s + "\n"
				}
			}
		}
		kml.Placemarks[i].Linestring.Coords = coordinates

	}
}
func (kml *Kml) ToXML() []byte {
	kml.ConvertCoords()
	doc := &File{Kml: *kml}
	xml, _ := xml.Marshal(doc)

	return xml
}
func (kml *Kml) SaveFile(file string) {
	kml.ConvertCoords()
	doc := &File{Kml: *kml}
	xml, _ := xml.Marshal(doc)
	//j :=  []byte(json)
	//log.Print(string(j))

	err := ioutil.WriteFile(file, xml, 0644)
	if err != nil {
		log.Print(err)
	}
}

func compareCoords(a, b [][]float64) bool {

	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for x := range a {
		for i := range a[x] {
			if a[x][i] != b[x][i] {
				return false
			}
		}
	}

	return true
}

func (kml *Kml) SortPlacemarkers() {
	trails := make(map[string][]Placemark)
	//sort.Slice(kml.Placemarks, func(i, j int) bool {
	//	return kml.Placemarks[i].Name > kml.Placemarks[j].Name
	//})
	//for _, v := range kml.Placemarks {
	//	log.Print(v.Name)
	//}
	for _, mark := range kml.Placemarks {
		if len(trails[mark.Name]) == 0 {
			trails[mark.Name] = []Placemark{mark}
		} else {
			trails[mark.Name] = append(trails[mark.Name], mark)
		}
	}

	//for k, _ := range trails {
	//	sort.Slice(trails[k], func(i, j int) bool {
	//		//name := trails[k][i].Name
	//		coords1 := trails[k][i].Linestring.Coordinates
	//		coords2 := trails[k][j].Linestring.Coordinates
	//		if coords1[0][1] > coords2[0][1] {
	//			return true
	//		} else {
	//			return false
	//		}
	//	})
	//}
	//var placemarks []Placemark
	//log.Print(trails["Catoctin National Recreation Trail"])

	for k, _ := range trails {
		//furthest := 0.0
		furthestId := ""
		if k == "Catoctin National Recreation Trail" {
			placemark := Lineup(trails[k])
			kml.Placemarks = []Placemark{placemark}
		}
		//for i, val := range trails[k] {
		//	//var placemark Placemark
		//	//newCoord := val.Linestring.Coordinates
		//	end_1 := val.Linestring.Coordinates[len(val.Linestring.Coordinates)-1]
		//	begin_1 := val.Linestring.Coordinates[0]
		//
		//	if val.Name == "Catoctin National Recreation Trail" {
		//		dis := distance(39.4917075, -77.4841673, 39.4992146, -77.4831109)
		//		log.Print("distance", dis)
		//		log.Print("index ", fmt.Sprintf("%v", i), "lat ", fmt.Sprintf("%v", val.Linestring.Coordinates[0][1]), " lon ", fmt.Sprintf("%v", val.Linestring.Coordinates[0][0]))
		//		log.Print("index ", fmt.Sprintf("%v", i), "lat ", fmt.Sprintf("%v", end_1[1]), " lon ", fmt.Sprintf("%v", end_1[0]))
		//
		//	}
		//	//begin_1 := val.Linestring.Coordinates[0]
		//	//
		//	//dir := ""
		//	underOne := false
		//	for _, mark := range trails[k] {
		//		begin_2 := mark.Linestring.Coordinates[0]
		//		end_2 := mark.Linestring.Coordinates[len(mark.Linestring.Coordinates)-1]
		//		if mark.Id != val.Id {
		//			d_start := distance(end_1[1], end_1[0], begin_2[1], begin_2[0])
		//			d_end := distance(begin_1[1], begin_1[0], end_2[1], end_2[0])
		//			d_start_to_start := distance(begin_1[1], begin_1[0], begin_2[1], begin_2[0])
		//			d_end_to_end := distance(end_1[1], end_1[0], end_2[1], end_2[0])
		//
		//			//log.Print(d_start_to_start)
		//			//log.Print(d_end_to_end)
		//			//d_end := distance(begin_1[1], begin_1[0], end_2[1], end_2[0])
		//			if d_start == 0 {
		//				log.Print("end to start")
		//				log.Print(val.Id, ",", mark.Id)
		//			}
		//			if d_end == 0 {
		//				log.Print("start to end")
		//				log.Print(val.Id, ",", mark.Id)
		//			}
		//			if d_start_to_start == 0 {
		//				log.Print("start to start")
		//				log.Print(val.Id, ",", mark.Id)
		//			}
		//			if d_end_to_end == 0 {
		//				log.Print("end to end")
		//				log.Print(val.Id, ",", mark.Id)
		//			}
		//			if d_start < 1 {
		//				underOne = true
		//				if val.Name == "Catoctin National Recreation Trail" {
		//					//log.Print(d_start)
		//					//log.Print(end_1[1], end_1[0], begin_2[1], begin_2[0])
		//				}
		//			}
		//
		//		}
		//	}
		//
		//	if underOne == false && val.Name == "Catoctin National Recreation Trail" {
		//		furthestId = val.Id
		//	}
		//	//placemark.Linestring.Coordinates = newCoord
		//	//placemarks = append(placemarks, placemark)
		//
		//}
		if trails[k][0].Name == "Catoctin National Recreation Trail" {
			log.Print("furthest")
			log.Print(furthestId)
		}
		//	newTrail := trails[k][0]
		//	var coords [][]float64
		//	for _, mark := range trails[k] {
		//			coords = append(coords,mark.Linestring.Coordinates...)
		//	}
		//	sort.Slice(coords, func(i, j int) bool {
		//		//name := trails[k][i].Name
		//		coords1 := coords[i]
		//		coords2 := coords[j]
		//		if coords1
		//		if d < .3 {
		//			return true
		//		}
		//		return false
		//	})
		//	newTrail.Linestring.Coordinates = coords
		//	placemarks = append(placemarks, newTrail)

		//var placemark Placemark
		//for i, mark := range trails[k] {
		//	if mark.Name == "Yellow Poplar Trail" {
		//		log.Print(mark.Linestring.Coordinates)
		//	}
		//	if i > 0 {
		//		placemark.Linestring.Coordinates = append(placemark.Linestring.Coordinates, mark.Linestring.Coordinates...)
		//	} else {
		//		placemark = mark
		//	}
		//}
		//
		//placemarks = append(placemarks, placemark)

	}

	//kml.Placemarks = placemarks

}

func distance(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64 {
	pos1 := haversine.Coord{Lat: lat1, Lon: lon1} //
	pos2 := haversine.Coord{Lat: lat2, Lon: lon2} //
	_, km := haversine.Distance(pos1, pos2)

	return km
}
func Lineup(placemarks []Placemark) Placemark {
	var newPlaces []Placemark
	var oldPlaces []Placemark

	log.Print("length ", len(placemarks))
	placemark := placemarks[0]
	for i, val := range placemarks {
		end_1 := val.Linestring.Coordinates[len(val.Linestring.Coordinates)-1]
		begin_1 := val.Linestring.Coordinates[0]
		var newCoords [][]float64
		for x, mark := range placemarks {
			log.Print(len(mark.Linestring.Coordinates))
			begin_2 := mark.Linestring.Coordinates[0]
			log.Print(len(mark.Linestring.Coordinates) - 1)
			end_2 := mark.Linestring.Coordinates[len(mark.Linestring.Coordinates)-1]
			if mark.Id != val.Id {
				d_start := distance(end_1[1], end_1[0], begin_2[1], begin_2[0])
				d_end := distance(begin_1[1], begin_1[0], end_2[1], end_2[0])
				d_start_to_start := distance(begin_1[1], begin_1[0], begin_2[1], begin_2[0])
				d_end_to_end := distance(end_1[1], end_1[0], end_2[1], end_2[0])

				if d_start == 0 {
					oldPlaces = append(placemarks[:x], placemarks[x+1:]...)
					newPlaces = append(placemarks[:i], placemarks[i+1:]...)
					newPlaces = append(newPlaces, oldPlaces...)
					newPlaces = unique(newPlaces)
					placemark.Linestring.Coordinates = newCoords
					newPlaces = append(newPlaces, placemark)
					return Lineup(newPlaces)
				} else if d_end == 0 {
					oldPlaces = append(placemarks[:x], placemarks[x+1:]...)
					newPlaces = append(placemarks[:i], placemarks[i+1:]...)
					newPlaces = append(newPlaces, oldPlaces...)
					newPlaces = unique(newPlaces)
					placemark.Linestring.Coordinates = newCoords
					newPlaces = append(newPlaces, placemark)
					return Lineup(newPlaces)
				} else if d_start_to_start == 0 {
					newLine := [][]float64{}
					for i := len(val.Linestring.Coordinates) - 1; i > 0; i-- {
						newLine = append(newLine, val.Linestring.Coordinates[i])
					}
					oldPlaces = append(placemarks[:x], placemarks[x+1:]...)
					newPlaces = append(placemarks[:i], placemarks[i+1:]...)
					newPlaces = append(newPlaces, oldPlaces...)
					newPlaces = unique(newPlaces)
					placemark.Linestring.Coordinates = newCoords
					newPlaces = append(newPlaces, placemark)
					return Lineup(newPlaces)

				} else if d_end_to_end == 0 {
					newLine := [][]float64{}
					for i := len(mark.Linestring.Coordinates) - 1; i > 0; i-- {
						newLine = append(newLine, mark.Linestring.Coordinates[i])
					}
					oldPlaces = append(placemarks[:x], placemarks[x+1:]...)
					newPlaces = append(placemarks[:i], placemarks[i+1:]...)
					newPlaces = append(newPlaces, oldPlaces...)
					newPlaces = unique(newPlaces)
					newCoords = append(val.Linestring.Coordinates, newLine...)
					placemark.Linestring.Coordinates = newCoords
					newPlaces = append(newPlaces, placemark)

					return Lineup(newPlaces)
				}
			}
		}
	}
	return placemark

}
func unique(floatSlice []Placemark) []Placemark {
	keys := make(map[string]bool)
	list := []Placemark{}
	for _, entry := range floatSlice {
		if _, value := keys[entry.Id]; !value {
			keys[entry.Id] = true
			list = append(list, entry)
		}
	}
	return list
}

func NewUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
