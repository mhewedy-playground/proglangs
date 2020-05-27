package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"strings"
)

type lang struct {
	name         string
	link         string
	influencedBy []*lang
	influenced   []*lang
}

func (l lang) String() string {
	return l.name
}

type langs []*lang

func (ll langs) String() string {
	var out = make([]string, len(ll))

	for i, l := range ll {
		out[i] = l.name
	}
	return strings.Join(out, ", ")
}

func (l *lang) get() error {

	res, err := http.Get("https://en.wikipedia.org" + l.link)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	infoBoxSelector := `#mw-content-text > div > table.infobox.vevent > tbody > tr`
	doc.Find(infoBoxSelector).Each(func(i int, s *goquery.Selection) {
		th := s.Find("th")
		if "Influenced by" == th.Text() {
			l.influencedBy = build(th)
		}
		if "Influenced" == th.Text() {
			l.influenced = build(th)
		}
	})

	return nil
}

func build(node *goquery.Selection) []*lang {
	var langs []*lang
	next := node.Parent().Next()

	for _, n := range next.Find("td > a").Nodes {
		langName := goquery.NewDocumentFromNode(n).Text()
		langs = append(langs, &lang{
			name: strings.ReplaceAll(langName, "(programming language)", ""),
			link: n.Attr[0].Val,
		})
	}
	return langs
}

func traverse(lang *lang) error {
	stmt := "💎 lang " + lang.String()

	if err := lang.get(); err != nil {
		return err
	}

	if lang.influencedBy != nil {
		stmt += fmt.Sprintf(" 🚀 InfluencedBy %s", langs(lang.influencedBy))

		for _, l := range lang.influencedBy {
			if err := traverse(l); err != nil {
				return err
			}
		}
	}

	if lang.influenced != nil {
		stmt += fmt.Sprintf(" 🚀 Influenced %s", langs(lang.influenced))

		for _, l := range lang.influenced {
			if err := traverse(l); err != nil {
				return err
			}
		}
	}

	stmt += "\n"
	fmt.Println(stmt)

	return nil
}

func main() {
	golang := &lang{
		name: "Go",
		link: "/wiki/Go_(programming_language)",
	}
	if err := traverse(golang); err != nil {
		log.Fatal(err)
	}
}
