[![GoDoc](https://godoc.org/github.com/eduncan911/podcast?status.svg)](https://godoc.org/github.com/eduncan911/podcast)
[![Build Status](https://travis-ci.org/eduncan911/podcast.svg?branch=master)](https://travis-ci.org/eduncan911/podcast)
[![Coverage Status](https://coveralls.io/repos/github/eduncan911/podcast/badge.svg?branch=master)](https://coveralls.io/github/eduncan911/podcast?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/eduncan911/podcast)](https://goreportcard.com/report/github.com/eduncan911/podcast)
[![MIT License](https://img.shields.io/npm/l/mediaelement.svg)](https://eduncan911.mit-license.org/)

# podcast
Package podcast generates a fully compliant iTunes and RSS 2.0 podcast feed
for GoLang using a simple API.

Full documentation with detailed examples located at <a href="https://godoc.org/github.com/eduncan911/podcast">https://godoc.org/github.com/eduncan911/podcast</a>

### Usage
To use, `go get` and `import` the package like your typical GoLang library.

	$ go get -u github.com/eduncan911/podcast
	
	import "github.com/eduncan911/podcast"

The API exposes a number of method receivers on structs that implements the
logic required to comply with the specifications and ensure a compliant feed.
A number of overrides occur to help with iTunes visibility of your episodes.

Notably, the [Podcast.AddItem(i Item)](#Podcast.AddItem) function performs most
of the heavy lifting by taking the [Item](#Item) input and performing
validation, overrides and duplicate setters through the feed.

See the detailed Examples of the API at <a href="https://godoc.org/github.com/eduncan911/podcast">https://godoc.org/github.com/eduncan911/podcast</a>.

### Extensibility
In no way are you restricted in having full control over your feeds.  You may
choose to skip the API methods and instead use the structs directly.  The
fields have been grouped by RSS 2.0 and iTunes fields.

iTunes specific fields are all prefixed with the letter `I`.

### References
RSS 2.0: <a href="https://cyber.harvard.edu/rss/rss.html">https://cyber.harvard.edu/rss/rss.html</a>

Podcasts: <a href="https://help.apple.com/itc/podcasts_connect/#/itca5b22233">https://help.apple.com/itc/podcasts_connect/#/itca5b22233</a>

### Release Notes
1.3.0
* fixes Item.Duration being set incorrectly.
* changed Item.AddEnclosure() parameter definition (Bytes not Seconds!).
* added Item.AddDuration formatting and override.
* added more documentation surrounding Item.Enclosure{}

1.2.1
* added Podcast.AddSubTitle() and truncating to 64 chars.
* added a number of Guards to protect against empty fields.

1.2.0
* added Podcast.AddPubDate() and Podcast.AddLastBuildDate() overrides.
* added Item.AddImage() to mask some cumbersome addition of IImage.
* added Item.AddPubDate to simply datetime setters.
* added more examples (mostly around Item struct).
* tweaked some documentation.

1.1.0
* Enabling CDATA in ISummary fields for Podcast and Channel.

1.0.0
* Initial release.
* Full documentation, full examples and complete code coverage.

## Table of Contents

* [Imported Packages](#pkg-imports)
* [Index](#pkg-index)
* [Examples](#pkg-examples)

## <a name="pkg-imports">Imported Packages</a>

- [github.com/pkg/errors](https://godoc.org/github.com/pkg/errors)

## <a name="pkg-index">Index</a>
* [type Author](#Author)
* [type Enclosure](#Enclosure)
* [type EnclosureType](#EnclosureType)
  * [func (et EnclosureType) String() string](#EnclosureType.String)
* [type ICategory](#ICategory)
* [type IImage](#IImage)
* [type ISummary](#ISummary)
* [type Image](#Image)
* [type Item](#Item)
  * [func (i \*Item) AddDuration(durationInSeconds int64)](#Item.AddDuration)
  * [func (i \*Item) AddEnclosure(url string, enclosureType EnclosureType, lengthInBytes int64)](#Item.AddEnclosure)
  * [func (i \*Item) AddImage(url string)](#Item.AddImage)
  * [func (i \*Item) AddPubDate(datetime \*time.Time)](#Item.AddPubDate)
  * [func (i \*Item) AddSummary(summary string)](#Item.AddSummary)
* [type Podcast](#Podcast)
  * [func New(title, link, description string, pubDate, lastBuildDate \*time.Time) Podcast](#New)
  * [func (p \*Podcast) AddAuthor(name, email string)](#Podcast.AddAuthor)
  * [func (p \*Podcast) AddCategory(category string, subCategories []string)](#Podcast.AddCategory)
  * [func (p \*Podcast) AddImage(url string)](#Podcast.AddImage)
  * [func (p \*Podcast) AddItem(i Item) (int, error)](#Podcast.AddItem)
  * [func (p \*Podcast) AddLastBuildDate(datetime \*time.Time)](#Podcast.AddLastBuildDate)
  * [func (p \*Podcast) AddPubDate(datetime \*time.Time)](#Podcast.AddPubDate)
  * [func (p \*Podcast) AddSubTitle(subTitle string)](#Podcast.AddSubTitle)
  * [func (p \*Podcast) AddSummary(summary string)](#Podcast.AddSummary)
  * [func (p \*Podcast) Bytes() []byte](#Podcast.Bytes)
  * [func (p \*Podcast) Encode(w io.Writer) error](#Podcast.Encode)
  * [func (p \*Podcast) String() string](#Podcast.String)
* [type TextInput](#TextInput)

#### <a name="pkg-examples">Examples</a>
* [Item.AddDuration](#example_Item_AddDuration)
* [Item.AddPubDate](#example_Item_AddPubDate)
* [New](#example_New)
* [Podcast.AddAuthor](#example_Podcast_AddAuthor)
* [Podcast.AddCategory](#example_Podcast_AddCategory)
* [Podcast.AddImage](#example_Podcast_AddImage)
* [Podcast.AddItem](#example_Podcast_AddItem)
* [Podcast.AddLastBuildDate](#example_Podcast_AddLastBuildDate)
* [Podcast.AddPubDate](#example_Podcast_AddPubDate)
* [Podcast.AddSummary](#example_Podcast_AddSummary)
* [Podcast.Bytes](#example_Podcast_Bytes)
* [Package (HttpHandlers)](#example__httpHandlers)
* [Package (IoWriter)](#example__ioWriter)

#### <a name="pkg-files">Package files</a>
[author.go](./author.go) [doc.go](./doc.go) [enclosure.go](./enclosure.go) [image.go](./image.go) [item.go](./item.go) [itunes.go](./itunes.go) [podcast.go](./podcast.go) [textinput.go](./textinput.go) 

## <a name="Author">type</a> [Author](./author.go#L8-L12)
``` go
type Author struct {
    XMLName xml.Name `xml:"itunes:owner"`
    Name    string   `xml:"itunes:name"`
    Email   string   `xml:"itunes:email"`
}
```
Author represents a named author and email.

For iTunes compliance, both Name and Email are required.

## <a name="Enclosure">type</a> [Enclosure](./enclosure.go#L46-L65)
``` go
type Enclosure struct {
    XMLName xml.Name `xml:"enclosure"`

    // URL is the downloadable url for the content. (Required)
    URL string `xml:"url,attr"`

    // Length is the size in Bytes of the download. (Required)
    Length int64 `xml:"-"`
    // LengthFormatted is the size in Bytes of the download. (Required)
    //
    // This field gets overwritten with the API when setting Length.
    LengthFormatted string `xml:"length,attr"`

    // Type is MIME type encoding of the download. (Required)
    Type EnclosureType `xml:"-"`
    // TypeFormatted is MIME type encoding of the download. (Required)
    //
    // This field gets overwritten with the API when setting Type.
    TypeFormatted string `xml:"type,attr"`
}
```
Enclosure represents a download enclosure.

## <a name="EnclosureType">type</a> [EnclosureType](./enclosure.go#L21)
``` go
type EnclosureType int
```
EnclosureType specifies the type of the enclosure.

``` go
const (
    M4A EnclosureType = iota
    M4V
    MP4
    MP3
    MOV
    PDF
    EPUB
)
```
EnclosureType specifies the type of the enclosure.

### <a name="EnclosureType.String">func</a> (EnclosureType) [String](./enclosure.go#L24)
``` go
func (et EnclosureType) String() string
```
String returns the MIME type encoding of the specified EnclosureType.

## <a name="ICategory">type</a> [ICategory](./itunes.go#L9-L13)
``` go
type ICategory struct {
    XMLName     xml.Name `xml:"itunes:category"`
    Text        string   `xml:"text,attr"`
    ICategories []*ICategory
}
```
ICategory is a 2-tier classification system for iTunes.

## <a name="IImage">type</a> [IImage](./itunes.go#L23-L26)
``` go
type IImage struct {
    XMLName xml.Name `xml:"itunes:image"`
    HREF    string   `xml:"href,attr"`
}
```
IImage represents an iTunes image.

Podcast feeds contain artwork that is a minimum size of
1400 x 1400 pixels and a maximum size of 3000 x 3000 pixels,
72 dpi, in JPEG or PNG format with appropriate file
extensions (.jpg, .png), and in the RGB colorspace. To optimize
images for mobile devices, Apple recommends compressing your
image files.

## <a name="ISummary">type</a> [ISummary](./itunes.go#L31-L34)
``` go
type ISummary struct {
    XMLName xml.Name `xml:"itunes:summary"`
    Text    string   `xml:",cdata"`
}
```
ISummary is a 4000 character rich-text field for the itunes:summary tag.

This is rendered as CDATA which allows for HTML tags such as <a href="">.

## <a name="Image">type</a> [Image](./image.go#L13-L21)
``` go
type Image struct {
    XMLName xml.Name `xml:"image"`
    // TODO: is it URL or Link? which is it?
    URL    string `xml:"url"`
    Title  string `xml:"title,omitempty"`
    Link   string `xml:"link,omitempty"`
    Width  int    `xml:"width,omitempty"`
    Height int    `xml:"height,omitempty"`
}
```
Image represents an image.

Podcast feeds contain artwork that is a minimum size of
1400 x 1400 pixels and a maximum size of 3000 x 3000 pixels,
72 dpi, in JPEG or PNG format with appropriate file
extensions (.jpg, .png), and in the RGB colorspace. To optimize
images for mobile devices, Apple recommends compressing your
image files.

## <a name="Item">type</a> [Item](./item.go#L25-L49)
``` go
type Item struct {
    XMLName          xml.Name   `xml:"item"`
    GUID             string     `xml:"guid"`
    Title            string     `xml:"title"`
    Link             string     `xml:"link"`
    Description      string     `xml:"description"`
    Author           *Author    `xml:"-"`
    AuthorFormatted  string     `xml:"author,omitempty"`
    Category         string     `xml:"category,omitempty"`
    Comments         string     `xml:"comments,omitempty"`
    Source           string     `xml:"source,omitempty"`
    PubDate          *time.Time `xml:"-"`
    PubDateFormatted string     `xml:"pubDate,omitempty"`
    Enclosure        *Enclosure

    // https://help.apple.com/itc/podcasts_connect/#/itcb54353390
    IAuthor            string `xml:"itunes:author,omitempty"`
    ISubtitle          string `xml:"itunes:subtitle,omitempty"`
    ISummary           *ISummary
    IImage             *IImage
    IDuration          string `xml:"itunes:duration,omitempty"`
    IExplicit          string `xml:"itunes:explicit,omitempty"`
    IIsClosedCaptioned string `xml:"itunes:isClosedCaptioned,omitempty"`
    IOrder             string `xml:"itunes:order,omitempty"`
}
```
Item represents a single entry in a podcast.

Article minimal requirements are:
- Title
- Description
- Link

Audio minimal requirements are:
- Title
- Description
- Enclosure (HREF, Type and Length all required)

Recommendations:
- Setting the minimal fields sets most of other fields, including iTunes.
- Use the Published time.Time setting instead of PubDate.
- Always set an Enclosure.Length, to be nice to your downloaders.
- Use Enclosure.Type instead of setting TypeFormatted for valid extensions.

### <a name="Item.AddDuration">func</a> (\*Item) [AddDuration](./item.go#L101)
``` go
func (i *Item) AddDuration(durationInSeconds int64)
```
AddDuration adds the duration to the iTunes duration field.

### <a name="Item.AddEnclosure">func</a> (\*Item) [AddEnclosure](./item.go#L52-L53)
``` go
func (i *Item) AddEnclosure(
    url string, enclosureType EnclosureType, lengthInBytes int64)
```
AddEnclosure adds the downloadable asset to the podcast Item.

### <a name="Item.AddImage">func</a> (\*Item) [AddImage](./item.go#L70)
``` go
func (i *Item) AddImage(url string)
```
AddImage adds the image as an iTunes-only IImage.  RSS 2.0 does not have
the specification of Images at the Item level.

Podcast feeds contain artwork that is a minimum size of
1400 x 1400 pixels and a maximum size of 3000 x 3000 pixels,
72 dpi, in JPEG or PNG format with appropriate file
extensions (.jpg, .png), and in the RGB colorspace. To optimize
images for mobile devices, Apple recommends compressing your
image files.

### <a name="Item.AddPubDate">func</a> (\*Item) [AddPubDate](./item.go#L79)
``` go
func (i *Item) AddPubDate(datetime *time.Time)
```
AddPubDate adds the datetime as a parsed PubDate.

UTC time is used by default.

### <a name="Item.AddSummary">func</a> (\*Item) [AddSummary](./item.go#L90)
``` go
func (i *Item) AddSummary(summary string)
```
AddSummary adds the iTunes summary.

Limit: 4000 characters

Note that this field is a CDATA encoded field which allows for rich text
such as html links: <a href="<a href="http://www.apple.com">http://www.apple.com</a>">Apple</a>.

## <a name="Podcast">type</a> [Podcast](./podcast.go#L19-L57)
``` go
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
    // contains filtered or unexported fields
}
```
Podcast represents a podcast.

### <a name="New">func</a> [New](./podcast.go#L63-L64)
``` go
func New(title, link, description string,
    pubDate, lastBuildDate *time.Time) Podcast
```
New instantiates a Podcast with required parameters.

Nil-able fields are optional but recommended as they are formatted
to the expected proper formats.

### <a name="Podcast.AddAuthor">func</a> (\*Podcast) [AddAuthor](./podcast.go#L80)
``` go
func (p *Podcast) AddAuthor(name, email string)
```
AddAuthor adds the specified Author to the podcast.

### <a name="Podcast.AddCategory">func</a> (\*Podcast) [AddCategory](./podcast.go#L91)
``` go
func (p *Podcast) AddCategory(category string, subCategories []string)
```
AddCategory adds the categories to the Podcast in comma delimited format.

subCategories are optional.

### <a name="Podcast.AddImage">func</a> (\*Podcast) [AddImage](./podcast.go#L122)
``` go
func (p *Podcast) AddImage(url string)
```
AddImage adds the specified Image to the Podcast.

Podcast feeds contain artwork that is a minimum size of
1400 x 1400 pixels and a maximum size of 3000 x 3000 pixels,
72 dpi, in JPEG or PNG format with appropriate file
extensions (.jpg, .png), and in the RGB colorspace. To optimize
images for mobile devices, Apple recommends compressing your
image files.

### <a name="Podcast.AddItem">func</a> (\*Podcast) [AddItem](./podcast.go#L173)
``` go
func (p *Podcast) AddItem(i Item) (int, error)
```
AddItem adds the podcast episode.  It returns a count of Items added or any
errors in validation that may have occurred.

This method takes the "itunes overrides" approach to populating
itunes tags according to the overrides rules in the specification.
This not only complies completely with iTunes parsing rules; but, it also
displays what is possible to be set on an individual episode level - if you
wish to have more fine grain control over your content.

This method imposes strict validation of the Item being added to confirm
to Podcast and iTunes specifications.

Article minimal requirements are:

* Title
* Description
* Link

Audio, Video and Downloads minimal requirements are:

* Title
* Description
* Enclosure (HREF, Type and Length all required)

The following fields are always overwritten (don't set them):

* GUID
* PubDateFormatted
* AuthorFormatted
* Enclosure.TypeFormatted
* Enclosure.LengthFormatted

Recommendations:

* Just set the minimal fields: the rest get set for you.
* Always set an Enclosure.Length, to be nice to your downloaders.
* Follow Apple's best practices to enrich your podcasts:

	<a href="https://help.apple.com/itc/podcasts_connect/#/itc2b3780e76">https://help.apple.com/itc/podcasts_connect/#/itc2b3780e76</a>

* For specifications of itunes tags, see:

	<a href="https://help.apple.com/itc/podcasts_connect/#/itcb54353390">https://help.apple.com/itc/podcasts_connect/#/itcb54353390</a>

### <a name="Podcast.AddLastBuildDate">func</a> (\*Podcast) [AddLastBuildDate](./podcast.go#L247)
``` go
func (p *Podcast) AddLastBuildDate(datetime *time.Time)
```
AddLastBuildDate adds the datetime as a parsed PubDate.

UTC time is used by default.

### <a name="Podcast.AddPubDate">func</a> (\*Podcast) [AddPubDate](./podcast.go#L240)
``` go
func (p *Podcast) AddPubDate(datetime *time.Time)
```
AddPubDate adds the datetime as a parsed PubDate.

UTC time is used by default.

### <a name="Podcast.AddSubTitle">func</a> (\*Podcast) [AddSubTitle](./podcast.go#L256)
``` go
func (p *Podcast) AddSubTitle(subTitle string)
```
AddSubTitle adds the iTunes subtitle that is displayed with the title
in iTunes.

Note that this field should be just a few words long according to Apple.
This method will truncate the string to 64 chars if too long with "..."

### <a name="Podcast.AddSummary">func</a> (\*Podcast) [AddSummary](./podcast.go#L273)
``` go
func (p *Podcast) AddSummary(summary string)
```
AddSummary adds the iTunes summary.

Limit: 4000 characters

Note that this field is a CDATA encoded field which allows for rich text
such as html links: <a href="<a href="http://www.apple.com">http://www.apple.com</a>">Apple</a>.

### <a name="Podcast.Bytes">func</a> (\*Podcast) [Bytes](./podcast.go#L287)
``` go
func (p *Podcast) Bytes() []byte
```
Bytes returns an encoded []byte slice.

### <a name="Podcast.Encode">func</a> (\*Podcast) [Encode](./podcast.go#L292)
``` go
func (p *Podcast) Encode(w io.Writer) error
```
Encode writes the bytes to the io.Writer stream in RSS 2.0 specification.

### <a name="Podcast.String">func</a> (\*Podcast) [String](./podcast.go#L303)
``` go
func (p *Podcast) String() string
```
String encodes the Podcast state to a string.

## <a name="TextInput">type</a> [TextInput](./textinput.go#L6-L12)
``` go
type TextInput struct {
    XMLName     xml.Name `xml:"textInput"`
    Title       string   `xml:"title"`
    Description string   `xml:"description"`
    Name        string   `xml:"name"`
    Link        string   `xml:"link"`
}
```
TextInput represents text inputs.

- - -
Generated by [godoc2ghmd](https://github.com/eduncan911/godoc2ghmd)