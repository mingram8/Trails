package openStreetMap

import (
	"encoding/xml"
	"log"
)

type Node struct {
	XMLName xml.Name `xml:"node"`
	Id      string   `xml:"id,attr"`
	Visible bool     `xml:"visible,attr"`
	Uid     int      `xml:"uid,attr"`
	Lat     float64  `xml:"lat,attr"`
	Lon     float64  `xml:"lon,attr"`
	Tags    Tag      `xml:"tag"`
	Name    string   `xml:"name,attr"`
	Wayid   string   `xml:"wayid"`
	Type    string   `xml:"type,attr"`
	Ski     Ski      `json:"ski"`
	Mtnbike Mtnbike  `json:"mtnbike"`
	Foot    Foot     `json:"foot"`
}
type Foot struct {
	Diff    string `json:"difficulty"`
	Tipo    string `json:"type"`
	Surface string `json:"surface"`
}
type Ski struct {
	Diff        string `json:"difficulty"`
	Description string `json:"description"`
	Tipo        string `json:"type"`
}
type Mtnbike struct {
	Diff        string `json:"difficulty"`
	Description string `json:"description"`
	Surface     string `json:"surface"`
}
type Osm struct {
	Ways  []Way  `xml:"way"`
	Nodes []Node `xml:"node"`
}
type Way struct {
	XMLName xml.Name `xml:"way"`
	Tags    []Tag    `xml:"tag"`
	Nds     []Nd     `xml:"nd"`
	Id      string   `xml:"id,attr"`
	Name    string   `xml:"name,attr"`
	Type    string   `xml:"type,attr"`
	Ski     Ski      `json:"ski"`
	Mtnbike Mtnbike  `json:"mtnbike"`
	Foot    Foot     `json:"foot"`
}

type Tag struct {
	XMLName xml.Name `xml:"tag"`
	Key     string   `xml:"k,attr"`
	Value   string   `xml:"v,attr"`
}
type Nd struct {
	XMLName xml.Name `xml:"nd"`
	Ref     string   `xml:"ref,attr"`
	Name    string   `xml:"name,attr"`
}

func CombineWays(way Way, way2 Way) (Way, bool) {
	nodes := way.Nds
	nodes_2 := way2.Nds

	for i, node := range nodes {
		for x, node_2 := range nodes_2 {
			if node.Ref == node_2.Ref && way.Name == way2.Name {
				log.Print(way2.Name)
				log.Print("node ", i, "node_2 ", x)
				log.Print(node.Ref)
				return way, true

			}
		}
	}
	return way, false
}
