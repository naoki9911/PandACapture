package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
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
			fmt.Println("HOGEHOGE")
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

var opts struct {
	ECSID        string `short:"i" long:"ecs-id" description:"ECS-ID to login" required:"true"`
	Password     string `short:"p" long:"password" description:"Password to login" required:"true"`
	DownloadPath string `short:"o" long:"output" description:"Path to download" default:"downloads"`
}

func main() {
	favorite := false
	sleepDur := 1 * time.Second

	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("DownloadPath: %s\n", opts.DownloadPath)

	h := panda.NewClient()
	err = h.Login(opts.ECSID, opts.Password)
	if err != nil {
		panic(err)
	}

	sites := []string{}
	if favorite {
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

	for i, siteID := range sites {
		fmt.Println(i)
		site, err := getSite(h, siteID)
		if err != nil {
			panic(err)
		}
		fmt.Println(site)
		if err != nil {
			fmt.Println(site)
			continue
		}

		err = site.download(h, opts.DownloadPath, sleepDur)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Download Done")
}
