package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/levigross/grequests"
	"github.com/tuotoo/biu/log"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type Movie struct {
	Title     string
	Link      string
	Desc      string
	CreatedAt string
	Download  string
}

func (m *Movie) Detail() error {
	if m.Link == "" {
		return fmt.Errorf("link is empty, %s", m.Title)
	}

	log.Info().Msgf("detail %s %s", m.Title, m.Link)
	resp, err := grequests.Get(m.Link, nil)
	if err != nil {
		return err
	}
	defer resp.Close()

	gbkResp := transform.NewReader(resp, simplifiedchinese.GBK.NewDecoder())

	doc, err := goquery.NewDocumentFromReader(gbkResp)
	if err != nil {
		return err
	}

	table := doc.Find("#downlist").First().Find("table").Children().First()
	if table != nil {
		a := table.Find("tr").Find("a")
		m.Download = a.AttrOr("href", "/")
		m.Download = strings.ReplaceAll(m.Download, "[电影天堂www.dytt89.com]", "")
	}

	log.Info().Msgf("-> %s", m.Download)
	return nil
}

const (
	host = "https://www.dy2018.com"
)

func main() {
	var path string
	var trims string
	flag.StringVar(&path, "output", "./rss.xml", "specify the output file path")
	flag.StringVar(&trims, "trims", "[电影天堂www.dytt89.com]", "specify the prefix to be trim")
	flag.Parse()

	movies, err := homepage()
	if err != nil {
		log.Fatal().Err(err)
	}

	for _, m := range movies {
		if err := m.Detail(); err != nil {
			log.Error().Err(err)
		}
	}

	rss, err := asRss(movies)
	if err != nil {
		log.Fatal().Err(err)
	}

	if err := ioutil.WriteFile(path, []byte(rss), 0755); err != nil {
		log.Fatal().Err(err)
	}
}

func homepage() ([]*Movie, error) {
	// 最新电影
	resp, err := grequests.Get(host+"/html/gndy/dyzz/index.html", nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := resp.Close()
		if err != nil {
			log.Info().Err(err).Msg("close resp")
		}
	}()

	gbkResp := transform.NewReader(resp, simplifiedchinese.GBK.NewDecoder())

	doc, err := goquery.NewDocumentFromReader(gbkResp)
	if err != nil {
		return nil, err
	}

	movies := make([]*Movie, 0)

	tables := doc.Find(".co_content8").First().Find("ul").Children()
	tables.Each(func(i int, table *goquery.Selection) {
		trs := table.Find("tr")
		a := trs.Eq(1).Find("a").First()
		td := trs.Eq(3).Find("td").First()
		movies = append(movies, &Movie{
			Title:     a.Text(),
			Link:      host + a.AttrOr("href", "/"),
			Desc:      td.Text(),
			CreatedAt: "",
			Download:  "",
		})
	})

	return movies, nil
}

func asRss(movies []*Movie) (string, error) {
	feed := &feeds.Feed{
		Title:  "电影天堂",
		Link:   &feeds.Link{Href: host},
		Author: &feeds.Author{Name: "Jqs7", Email: "7@jqs7.com"},
	}

	for _, m := range movies {
		i := &feeds.Item{
			Title:       m.Title,
			Link:        &feeds.Link{Href: m.Download},
			Description: m.Desc,
		}
		feed.Add(i)
	}
	return feed.ToRss()
}
