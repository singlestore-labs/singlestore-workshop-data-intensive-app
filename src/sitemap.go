package src

import (
	"math/rand"
	"net/url"
	"strings"

	sitemap "github.com/oxffaa/gopher-parse-sitemap"
)

type Page struct {
	Path     string
	Parent   *Page
	Children []*Page
}

func (p *Page) NewChild(path string) *Page {
	child := &Page{Path: path, Parent: p}
	p.Children = append(p.Children, child)
	return child
}

func (p *Page) AddChildRecursive(path string) {
	ptr := p
	totalPath := ""
	path = strings.TrimLeft(path, "/")
	for _, part := range strings.Split(path, "/") {
		totalPath += strings.TrimRight("/"+part, "/")
		ptr = ptr.NewChild(totalPath)
	}
}

// Get a random leaf node starting at this point on the tree
func (p *Page) RandomLeaf() *Page {
	if len(p.Children) == 0 {
		return p
	}

	return p.RandomChild().RandomLeaf()
}

// Get a random child node of this page, or return this page if it has no children
func (p *Page) RandomChild() *Page {
	if len(p.Children) == 0 {
		return p
	}

	return p.Children[rand.Intn(len(p.Children))]
}

// Get a random child or this pages parent, prefers to traverse down the tree
func (p *Page) RandomNext() *Page {
	if len(p.Children) > 0 {
		return p.RandomChild()
	}
	return p.Parent
}

func LoadSitemap(sitemapURL string) (*Page, error) {
	root := &Page{}

	return root, sitemap.ParseFromSite(sitemapURL, func(e sitemap.Entry) error {
		loc, err := url.Parse(e.GetLocation())
		if err != nil {
			return err
		}

		root.AddChildRecursive(loc.Path)

		return nil
	})
}
