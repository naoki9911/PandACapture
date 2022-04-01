package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/naoki9911/panda"
	"github.com/ssoroka/slice"
)

type Content struct {
	fileName string
	fileURL  string
	filePath string
}

type Site struct {
	name  string
	root  string
	files []Content
}

func newSite() *Site {
	return &Site{
		files: []Content{},
	}
}

func getSiteName(contents []panda.Content) (string, error) {
	for _, content := range contents {
		if content.Container == "/content/group/" {
			return content.EntityTitle, nil
		}
	}

	return "", fmt.Errorf("failed to get site name")
}

func getCollections(contents []panda.Content) []string {
	cols := []string{}
	for _, content := range contents {
		if content.Type == "collection" {
			continue
		}
		if slice.Contains(cols, content.Container) {
			continue
		}
		cols = append(cols, content.Container)
	}

	return slice.SortBy(cols, func(slice []string, i int, j int) bool { return len(slice[i]) < len(slice[j]) })
}

func getSite(handler *panda.Handler, siteID string) (*Site, error) {
	site := newSite()
	contents := handler.GetContent(siteID)

	siteName, err := getSiteName(contents)
	if err != nil {
		return nil, err
	}
	site.name = siteName

	collections := getCollections(contents)

	if len(collections) == 0 {
		return nil, fmt.Errorf("no contents")
	}
	// extract /content/group/13f35cd6-1440-4dcf-9395-d5d7c44c05d7/
	roots := strings.Split(collections[0], "/")
	if len(roots) < 4 {
		return nil, fmt.Errorf("%v", roots)
	}
	site.root = "/" + roots[1] + "/" + roots[2] + "/" + roots[3]

	for _, content := range contents {
		// ignore 'collection'
		if content.Type == "collection" {
			continue
		}

		dirName := strings.Replace(content.Container, site.root, "", 1)
		c := Content{
			fileName: content.Title,
			fileURL:  content.URL,
		}
		if filepath.Ext(c.fileName) == "" {
			ext, _ := slice.Last(strings.Split(c.fileURL, "."))
			c.fileName = c.fileName + "." + ext
		}

		c.filePath = path.Join(site.name, path.Join(dirName, c.fileName))
		site.files = append(site.files, c)
	}

	return site, nil
}

func createDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0744)
		if err != nil {
			return err
		}
	}
	return err
}

func (c *Content) download(h *panda.Handler, downloadDir string) error {
	filePath := path.Join(downloadDir, c.filePath)

	err := createDir(path.Dir(filePath))
	if err != nil {
		return err
	}

	fmt.Printf("Downloading %s\n", filePath)
	err = h.Download(filePath, c.fileURL)
	if err != nil {
		return err
	}

	return nil
}

func (site *Site) download(h *panda.Handler, downloadDir string, sleepDur time.Duration) error {
	for _, file := range site.files {
		err := file.download(h, downloadDir)
		if err != nil {
			fmt.Printf("failed to download %s\n", file.filePath)
		}

		// Sleep for avoiding DoS attack
		time.Sleep(sleepDur)
	}
	return nil
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-output DIR] [-favorite] [-sleep SECONDS] ECS-ID PASSWORD\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  -favorite\n")
	fmt.Fprintf(os.Stderr, "  	If true, only files marked 'favorite' are downloaded\n")
	fmt.Fprintf(os.Stderr, "  -output string\n")
	fmt.Fprintf(os.Stderr, "  	Path to store downloaded files (default \"downloads\")\n")
	fmt.Fprintf(os.Stderr, "  -sleep float\n")
	fmt.Fprintf(os.Stderr, "  	Durtaion (second) to sleep after downloading each file (default 1)\n")
}

func main() {
	var err error

	// Parse command-line arguments
	var (
		downloadPathOpt   = flag.String("output", "downloads", "")
		onlyFavoriteOpt   = flag.Bool("favorite", false, "")
		sleepDurSecondOpt = flag.Float64("sleep", 1, "")
	)
	flag.Usage = printUsage
	flag.Parse()
	if flag.NArg() != 2 {
		printUsage()
		os.Exit(1)
	}
	ecsId := flag.Arg(0)
	ecsPassword := flag.Arg(1)
	downloadPath := *downloadPathOpt
	onlyFavorite := *onlyFavoriteOpt
	sleepDur := time.Duration(*sleepDurSecondOpt * float64(time.Second))

	fmt.Printf("Downloaded files will be stored in: %s\n", downloadPath)

	// Create the client
	h := panda.NewClient()
	err = h.Login(ecsId, ecsPassword)
	if err != nil {
		panic(err)
	}

	// Download the files
	sites := []string{}
	if onlyFavorite {
		resp := h.GetFavoriteSites()
		for _, s := range resp.FavoriteSitesIDs {
			sites = append(sites, s)
		}
	} else {
		resp := h.GetAllSites()
		for _, s := range resp.SiteCollection {
			sites = append(sites, s.Id)
		}
	}

	for _, siteID := range sites {
		site, err := getSite(h, siteID)
		if err != nil {
			panic(err)
		}
		if err != nil {
			fmt.Println(site)
			continue
		}

		err = site.download(h, downloadPath, sleepDur)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Download Done")
}
