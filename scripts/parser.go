package main

import (
	"log"

	"github.com/comov/hsearch/parser"
)

func main() {
	//var site = parser.DieselSite()
	var site = parser.HouseSite()
	//var site = parser.LalafoSite()

	//doc, err := parser.GetDocumentByUrl(site.Url())
	//if err != nil {
	//	log.Fatalln(err)
	//}

	apartmentsLinks, err := parser.FindApartmentsLinksOnSite(site)
	if err != nil {
		log.Fatalln(err)
	}

	for id, link := range apartmentsLinks {
		log.Println(id, link)
	}
}
