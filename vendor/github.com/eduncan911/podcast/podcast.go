package podcast

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const (
	pVersion = "1.3.0"
)

// Podcast represents a podcast.
type Podcast struct {
	XMLName        xml.Name `xml:"channel"`
	Title          string   `xml:"title"`
	Link           string   `xml:"link"`
	Description    string   `xml:"description"`
	Category       string   `xml:"category,omitempty"`
	Cloud          string   `xml:"cloud,omitempty"`
	Copyright      string   `xml:"copyright,omitempty"`
	Docs           string   `xml:"docs,omitempty"`
	Generator      string   `xml:"generator,omitempty"`
	Language       string   `xml:"language,omitempty"`
	LastBuildDate  string   `xml:"lastBuildDate,omitempty"`
	ManagingEditor string   `xml:"managingEditor,omitempty"`
	PubDate        string   `xml:"pubDate,omitempty"`
	Rating         string   `xml:"rating,omitempty"`
	SkipHours      string   `xml:"skipHours,omitempty"`
	SkipDays       string   `xml:"skipDays,omitempty"`
	TTL            int      `xml:"ttl,omitempty"`
	WebMaster      string   `xml:"webMaster,omitempty"`
	Image          *Image
	TextInput      *TextInput

	// https://help.apple.com/itc/podcasts_connect/#/itcb54353390
	IAuthor     string `xml:"itunes:author,omitempty"`
	ISubtitle   string `xml:"itunes:subtitle,omitempty"`
	ISummary    *ISummary
	IBlock      string `xml:"itunes:block,omitempty"`
	IImage      *IImage
	IDuration   string  `xml:"itunes:duration,omitempty"`
	IExplicit   string  `xml:"itunes:explicit,omitempty"`
	IComplete   string  `xml:"itunes:complete,omitempty"`
	INewFeedURL string  `xml:"itunes:new-feed-url,omitempty"`
	IOwner      *Author // Author is formatted for itunes as-is
	ICategories []*ICategory

	Items []*Item

	encode func(w io.Writer, o interface{}) error
}

// New instantiates a Podcast with required parameters.
//
// Nil-able fields are optional but recommended as they are formatted
// to the expected proper formats.
func New(title, link, description string,
	pubDate, lastBuildDate *time.Time) Podcast {
	return Podcast{
		Title:         title,
		Link:          link,
		Description:   description,
		Generator:     fmt.Sprintf("go podcast v%s (github.com/eduncan911/podcast)", pVersion),
		PubDate:       parseDateRFC1123Z(pubDate),
		LastBuildDate: parseDateRFC1123Z(lastBuildDate),
		Language:      "en-us",

		// setup dependency (could inject later)
		encode: encoder,
	}
}

// AddAuthor adds the specified Author to the podcast.
func (p *Podcast) AddAuthor(name, email string) {
	p.ManagingEditor = parseAuthorNameEmail(&Author{
		Name:  name,
		Email: email,
	})
	p.IAuthor = p.ManagingEditor
}

// AddCategory adds the categories to the Podcast in comma delimited format.
//
// subCategories are optional.
func (p *Podcast) AddCategory(category string, subCategories []string) {
	if len(category) == 0 {
		return
	}

	// RSS 2.0 Category only supports 1-tier
	if len(p.Category) > 0 {
		p.Category = p.Category + "," + category
	} else {
		p.Category = category
	}

	icat := ICategory{Text: category}
	for _, c := range subCategories {
		if len(c) == 0 {
			continue
		}
		icat2 := ICategory{Text: c}
		icat.ICategories = append(icat.ICategories, &icat2)
	}
	p.ICategories = append(p.ICategories, &icat)
}

// AddImage adds the specified Image to the Podcast.
//
// Podcast feeds contain artwork that is a minimum size of
// 1400 x 1400 pixels and a maximum size of 3000 x 3000 pixels,
// 72 dpi, in JPEG or PNG format with appropriate file
// extensions (.jpg, .png), and in the RGB colorspace. To optimize
// images for mobile devices, Apple recommends compressing your
// image files.
func (p *Podcast) AddImage(url string) {
	if len(url) == 0 {
		return
	}
	p.Image = &Image{
		URL: url,
	}
	p.IImage = &IImage{HREF: url}
}

// AddItem adds the podcast episode.  It returns a count of Items added or any
// errors in validation that may have occurred.
//
// This method takes the "itunes overrides" approach to populating
// itunes tags according to the overrides rules in the specification.
// This not only complies completely with iTunes parsing rules; but, it also
// displays what is possible to be set on an individual episode level - if you
// wish to have more fine grain control over your content.
//
// This method imposes strict validation of the Item being added to confirm
// to Podcast and iTunes specifications.
//
// Article minimal requirements are:
//
// * Title
// * Description
// * Link
//
// Audio, Video and Downloads minimal requirements are:
//
// * Title
// * Description
// * Enclosure (HREF, Type and Length all required)
//
// The following fields are always overwritten (don't set them):
//
// * GUID
// * PubDateFormatted
// * AuthorFormatted
// * Enclosure.TypeFormatted
// * Enclosure.LengthFormatted
//
// Recommendations:
//
// * Just set the minimal fields: the rest get set for you.
// * Always set an Enclosure.Length, to be nice to your downloaders.
// * Follow Apple's best practices to enrich your podcasts:
//   https://help.apple.com/itc/podcasts_connect/#/itc2b3780e76
// * For specifications of itunes tags, see:
//   https://help.apple.com/itc/podcasts_connect/#/itcb54353390
//
func (p *Podcast) AddItem(i Item) (int, error) {
	// initial guards for required fields
	if len(i.Title) == 0 || len(i.Description) == 0 {
		return len(p.Items), errors.New("Title and Description are reuired")
	}
	if i.Enclosure != nil {
		if len(i.Enclosure.URL) == 0 {
			return len(p.Items),
				errors.New(i.Title + ": Enclosure.URL is required")
		}
		if i.Enclosure.Type.String() == enclosureDefault {
			return len(p.Items),
				errors.New(i.Title + ": Enclosure.Type is required")
		}
	} else if len(i.Link) == 0 {
		return len(p.Items),
			errors.New(i.Title + ": Link is required when not using Enclosure")
	}

	// corrective actions and overrides
	//
	i.PubDateFormatted = parseDateRFC1123Z(i.PubDate)
	i.AuthorFormatted = parseAuthorNameEmail(i.Author)
	if i.Enclosure != nil {
		i.GUID = i.Enclosure.URL // yep, GUID is the Permlink URL

		if i.Enclosure.Length < 0 {
			i.Enclosure.Length = 0
		}
		i.Enclosure.LengthFormatted = strconv.FormatInt(i.Enclosure.Length, 10)
		i.Enclosure.TypeFormatted = i.Enclosure.Type.String()

		// allow Link to be set for article references to Downloads,
		// otherwise set it to the enclosurer's URL.
		if len(i.Link) == 0 {
			i.Link = i.Enclosure.URL
		}
	} else {
		i.GUID = i.Link // yep, GUID is the Permlink URL
	}

	// iTunes it
	//
	if len(i.IAuthor) == 0 {
		if i.Author != nil {
			i.IAuthor = i.Author.Email
		} else if len(p.IAuthor) != 0 {
			i.Author = &Author{Email: p.IAuthor}
			i.IAuthor = p.IAuthor
		} else if len(p.ManagingEditor) != 0 {
			i.Author = &Author{Email: p.ManagingEditor}
			i.IAuthor = p.ManagingEditor
		}
	}
	if i.IImage == nil {
		if p.Image != nil {
			i.IImage = &IImage{HREF: p.Image.URL}
		}
	}

	p.Items = append(p.Items, &i)
	return len(p.Items), nil
}

// AddPubDate adds the datetime as a parsed PubDate.
//
// UTC time is used by default.
func (p *Podcast) AddPubDate(datetime *time.Time) {
	p.PubDate = parseDateRFC1123Z(datetime)
}

// AddLastBuildDate adds the datetime as a parsed PubDate.
//
// UTC time is used by default.
func (p *Podcast) AddLastBuildDate(datetime *time.Time) {
	p.LastBuildDate = parseDateRFC1123Z(datetime)
}

// AddSubTitle adds the iTunes subtitle that is displayed with the title
// in iTunes.
//
// Note that this field should be just a few words long according to Apple.
// This method will truncate the string to 64 chars if too long with "..."
func (p *Podcast) AddSubTitle(subTitle string) {
	if len(subTitle) == 0 {
		return
	}
	if len(subTitle) > 64 {
		s := []rune(subTitle)
		subTitle = string(s[0:61]) + "..."
	}
	p.ISubtitle = subTitle
}

// AddSummary adds the iTunes summary.
//
// Limit: 4000 characters
//
// Note that this field is a CDATA encoded field which allows for rich text
// such as html links: <a href="http://www.apple.com">Apple</a>.
func (p *Podcast) AddSummary(summary string) {
	if len(summary) == 0 {
		return
	}
	if len(summary) > 4000 {
		s := []rune(summary)
		summary = string(s[0:4000])
	}
	p.ISummary = &ISummary{
		Text: summary,
	}
}

// Bytes returns an encoded []byte slice.
func (p *Podcast) Bytes() []byte {
	return []byte(p.String())
}

// Encode writes the bytes to the io.Writer stream in RSS 2.0 specification.
func (p *Podcast) Encode(w io.Writer) error {
	w.Write([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"))
	wrapped := podcastWrapper{
		XMLNS:   "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Version: "2.0",
		Channel: p,
	}
	return p.encode(w, wrapped)
}

// String encodes the Podcast state to a string.
func (p *Podcast) String() string {
	b := new(bytes.Buffer)
	if err := p.Encode(b); err != nil {
		return "String: podcast.write returned the error: " + err.Error()
	}
	return b.String()
}

// // Write implements the io.Writer interface to write an RSS 2.0 stream
// // that is compliant to the RSS 2.0 specification.
// func (p *Podcast) Write(b []byte) (n int, err error) {
// 	buf := bytes.NewBuffer(b)
// 	if err := p.Encode(buf); err != nil {
// 		return 0, errors.Wrap(err, "Write: podcast.encode returned error")
// 	}
// 	return buf.Len(), nil
// }

type podcastWrapper struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	XMLNS   string   `xml:"xmlns:itunes,attr"`
	Channel *Podcast
}

var encoder = func(w io.Writer, o interface{}) error {
	e := xml.NewEncoder(w)
	e.Indent("", "  ")
	if err := e.Encode(o); err != nil {
		return errors.Wrap(err, "podcast.encode: Encode returned error")
	}
	return nil
}

var parseDateRFC1123Z = func(t *time.Time) string {
	if t != nil && !t.IsZero() {
		return t.Format(time.RFC1123Z)
	}
	return time.Now().UTC().Format(time.RFC1123Z)
}

var parseAuthorNameEmail = func(a *Author) string {
	var author string
	if a != nil {
		author = a.Email
		if len(a.Name) > 0 {
			author = fmt.Sprintf("%s (%s)", a.Email, a.Name)
		}
	}
	return author
}

var parseDuration = func(duration int64) string {
	// TODO: parse the output into iTunes nicely formatted version.
	//
	// iTunes supports the following:
	//  HH:MM:SS
	//  H:MM:SS
	//  MM:SS
	//  M:SS
	return strconv.FormatInt(duration, 10)
}
