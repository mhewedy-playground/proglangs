package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

var langMap = make(map[string]*lang)

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

func (l *lang) traverse() error {

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
		var l *lang
		langName := goquery.NewDocumentFromNode(n).Text()
		// comment contains check in order to avoid deep links between objects
		if contains(langMap, langName) {
			l = langMap[langName]
		} else {
			l = &lang{
				name: strings.ReplaceAll(langName, "(programming language)", ""),
				link: n.Attr[0].Val,
			}
		}
		langs = append(langs, l)
	}
	return langs
}

func traverse(lang *lang) {
	if contains(langMap, lang.name) {
		return
	}

	langMap[lang.name] = lang

	stmt := "💎 lang: " + lang.String()

	if err := lang.traverse(); err != nil {
		stmt += " ⚠️ (error during traversing)"
	}

	if lang.influencedBy != nil {
		stmt += fmt.Sprintf(" 🚀 Influenced By: %s", langs(lang.influencedBy))

		for _, l := range lang.influencedBy {
			traverse(l)
		}
	}

	if lang.influenced != nil {
		stmt += fmt.Sprintf(" 🎈 Influences: %s", langs(lang.influenced))

		for _, l := range lang.influenced {
			traverse(l)
		}
	}

	fmt.Println(stmt)
}

func main() {
	golang := &lang{
		name: "Go",
		link: "/wiki/Go_(programming_language)",
	}
	traverse(golang)

	select {}
}

func contains(ss map[string]*lang, s string) bool {
	for k := range ss {
		if k == s {
			return true
		}
	}
	return false
}
