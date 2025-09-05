package airtable

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	Client  HTTPClient
	attempt int
)

func init() {
	Client = &http.Client{
		Timeout: time.Second * 30,
	}
}

const (
	apiUrl = "https://api.airtable.com/v0"
)

type Airtable struct {
	apiKey        string
	xClientSecret string // metadata API
	base          string
	debug         bool
}

// New creates a new Airtable client.
// - apiKey: your API key
// - base: the base to use
func New(apiKey, base string, debug bool) *Airtable {
	return &Airtable{
		apiKey: apiKey,
		base:   base,
		debug:  debug,
	}
}

type Parameters struct {
	Name                  string   `json:"name"`                  // table name
	MaxRecords            string   `json:"maxRecords"`            // The maximum total number of records that will be returned in your requests. If this value is larger than pageSize (which is 100 by default), you may have to load multiple pages to reach this total. See the Pagination section below for more.
	PageSize              string   `json:"pageSize"`              // The number of records returned in each request. Must be less than or equal to 100. Default is 100. See the Pagination section below for more.
	View                  string   `json:"view"`                  // The name or ID of a view in the table. If set, only the records in that view will be returned. The records will be sorted according to the order of the view unless the sort parameter is included, which overrides that order. Fields hidden in this view will be returned in the results. To only return a subset of fields, use the fields parameter.
	Fields                []string `json:"fields"`                // Only data for fields whose names are in this list will be included in the result. If you don't need every field, you can use this parameter to reduce the amount of data transferred.
	UserLocale            Local    `json:"userLocale"`            // The user locale that should be used to format dates when using string as the cellFormat. This parameter is required when using string as the cellFormat.
	TimeZone              TimeZone `json:"timeZone"`              // The time zone that should be used to format dates when using string as the cellFormat. This parameter is required when using string as the cellFormat.
	FilterByFormula       string   `json:"filterByFormula"`       // A formula used to filter records. The formula will be evaluated for each record, and if the result is not 0, false, "", NaN, [], or #Error! the record will be included in the response. If combined with the view parameter, only records in that view which satisfy the formula will be returned.https://support.airtable.com/hc/en-us/articles/203255215-Formula-Field-Reference
	Sort                  []Sort   `json:"sort"`                  // A list of sort objects that specifies how the records will be ordered. Each sort object must have a field key specifying the name of the field to sort on, and an optional direction key that is either "asc" or "desc". The default direction is "asc".
	ReturnFieldsByFieldId string   `json:"returnFieldsByFieldId"` // An optional boolean value that lets you return field objects where the key is the field id. This defaults to false, which returns field objects where the key is the field name.
	Offset                string   `json:"offset"`                // The server returns one page of records at a time. Each page will contain pageSize records, which is 100 by default. If there are more records, the response will contain an offset. To fetch the next page of records, include offset in the next request's parameters. Pagination will stop when you've reached the end of your table. If the maxRecords parameter is passed, pagination will stop once you've reached this maximum.
}

type Sort struct {
	Field     string
	Direction SortDirection
}

type Options struct {
	IsReversed              bool   `json:"isReversed"`
	InverseLinkFieldID      string `json:"inverseLinkFieldId"`
	LinkedTableID           string `json:"linkedTableId"`
	PrefersSingleRecordLink bool   `json:"prefersSingleRecordLink"`
}

type Fields struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Type        string  `json:"type"`
	Options     Options `json:"options,omitempty"`
}

type Views struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type Tables struct {
	Tables []Table `json:"tables"`
}

type Table struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	PrimaryFieldID string   `json:"primaryFieldId"`
	Fields         []Fields `json:"fields"`
	Views          []Views  `json:"views"`
}

type Bases struct {
	Bases []Base `json:"bases"`
}

type Base struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	PermissionLevel string `json:"permissionLevel"`
}

// SetXAPIKey sets the X-API-Key header.
// - secret: the secret to use metadata API: https://airtable.com/api/meta
func (a *Airtable) SetXAPIKey(secret string) {
	a.xClientSecret = secret
}

func (a *Airtable) ListBases() (Bases, error) {
	var bases Bases
	p := url.URL{Path: "meta/bases"}
	err := a.call(GET, &p, nil, &bases)
	return bases, err
}

func (a *Airtable) BaseSchema(baseID string) (Tables, error) {
	var schema Tables
	p := url.URL{Path: fmt.Sprintf("meta/bases/%s/tables", baseID)}
	err := a.call(GET, &p, nil, &schema)
	return schema, err
}

func (a *Airtable) List(p Parameters, response interface{}) error {
	if p.Name == "" {
		return fmt.Errorf("table name is required")
	}

	values := url.Values{}
	if p.MaxRecords == "" {
		p.MaxRecords = "100"
	}

	if p.PageSize == "" {
		p.PageSize = "100"
	}

	if p.Offset != "" {
		values.Add("offset", p.Offset)
	}

	if p.UserLocale != "" {
		values.Add("userLocale", string(p.UserLocale))
	}
	if p.TimeZone != "" {
		values.Add("timeZone", string(p.TimeZone))
	}
	if p.ReturnFieldsByFieldId != "" {
		values.Add("returnFieldsByFieldId", p.ReturnFieldsByFieldId)
	}

	values.Add("maxRecords", p.MaxRecords)
	values.Add("pageSize", p.PageSize)
	values.Add("view", p.View)

	for _, f := range p.Fields {
		values.Add("fields[]", f)
	}

	for k, s := range p.Sort {
		values.Add(fmt.Sprintf("sort[%v][field]", k), s.Field)
		values.Add(fmt.Sprintf("sort[%v][direction]", k), string(s.Direction))
	}

	if p.FilterByFormula != "" {
		values.Add("filterByFormula", p.FilterByFormula)
	}

	path := url.URL{
		Path:     fmt.Sprintf("%s/%s", a.base, p.Name),
		RawQuery: values.Encode(),
	}

	return a.call(GET, &path, nil, response)
}

func (a *Airtable) Get(p Parameters, id string, response interface{}) error {
	if p.Name == "" {
		return fmt.Errorf("table name is required")
	}

	values := url.Values{}
	if p.UserLocale != "" {
		values.Add("userLocale", string(p.UserLocale))
	}
	if p.TimeZone != "" {
		values.Add("timeZone", string(p.TimeZone))
	}
	if p.ReturnFieldsByFieldId != "" {
		values.Add("returnFieldsByFieldId", p.ReturnFieldsByFieldId)
	}

	path := &url.URL{
		Path:     fmt.Sprintf("%s/%s/%s", a.base, p.Name, id),
		RawQuery: values.Encode(),
	}

	return a.call(GET, path, nil, response)
}

func (a *Airtable) Create(p Parameters, data []byte, response interface{}) error {
	if p.Name == "" {
		return fmt.Errorf("table name is required")
	}

	values := url.Values{}
	if p.UserLocale != "" {
		values.Add("userLocale", string(p.UserLocale))
	}
	if p.TimeZone != "" {
		values.Add("timeZone", string(p.TimeZone))
	}
	if p.ReturnFieldsByFieldId != "" {
		values.Add("returnFieldsByFieldId", p.ReturnFieldsByFieldId)
	}

	path := url.URL{
		Path:     fmt.Sprintf("%s/%s", a.base, p.Name),
		RawQuery: values.Encode(),
	}
	return a.call(POST, &path, data, response)
}

func (a *Airtable) Update(p Parameters, id string, data []byte, response interface{}) error {
	if p.Name == "" {
		return fmt.Errorf("table name is required")
	}

	values := url.Values{}
	if p.UserLocale != "" {
		values.Add("userLocale", string(p.UserLocale))
	}
	if p.TimeZone != "" {
		values.Add("timeZone", string(p.TimeZone))
	}
	if p.ReturnFieldsByFieldId != "" {
		values.Add("returnFieldsByFieldId", p.ReturnFieldsByFieldId)
	}

	path := url.URL{
		Path:     fmt.Sprintf("%s/%s/%s", a.base, p.Name, id),
		RawQuery: values.Encode(),
	}
	return a.call(PATCH, &path, data, response)
}

func (a *Airtable) Delete(p Parameters, id string) error {
	if p.Name == "" {
		return fmt.Errorf("table name is required")
	}
	path := url.URL{
		Path: fmt.Sprintf("%s/%s/%s", a.base, p.Name, id),
	}
	return a.call(DELETE, &path, nil, nil)
}

type methodHttp string

const (
	GET    methodHttp = http.MethodGet
	POST   methodHttp = http.MethodPost
	PUT    methodHttp = http.MethodPut
	PATCH  methodHttp = http.MethodPatch
	DELETE methodHttp = http.MethodDelete
)

type GeneralError struct {
	Error string `json:"error"`
}

type DetailedError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func decodeJSONError(response *http.Response) string {
	var err error

	bodyBytes := make([]byte, 1024*64) // being WAAAAY safe allowing for 64K
	response.Body.Read(bodyBytes)
	bodyBytes = bytes.Trim(bodyBytes, "\x00")

	var detailedErr DetailedError
	err = json.Unmarshal(bodyBytes, &detailedErr)
	if err == nil {
		return detailedErr.Error.Message + ", " + detailedErr.Error.Type
	}

	var generalErr GeneralError
	err = json.Unmarshal(bodyBytes, &generalErr)
	if err == nil {
		return generalErr.Error
	}

	return string(bodyBytes)
}

func (a *Airtable) call(method methodHttp, path *url.URL, payload []byte, response interface{}) error {
	req, _ := http.NewRequest(string(method), apiUrl+"/"+path.String(), bytes.NewBuffer(payload))

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.apiKey))
	if a.xClientSecret != "" {
		req.Header.Add("X-Airtable-Client-Secret", a.xClientSecret)
	}
	req.Header.Add("Content-Type", "application/json")

	if a.debug {
		dump, _ := httputil.DumpRequest(req, true)
		log.Println(string(dump))
	}

	res, err := Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests {
		if attempt < 5 {
			attempt++
			time.Sleep(time.Second * 1)
			return a.call(method, path, payload, response)
		}
		return fmt.Errorf("the API is limited to 5 requests per second per base. If you exceed this rate, you will receive a 429 status code and will need to wait 30 seconds before subsequent requests will succeed")
	}

	if res.StatusCode == http.StatusBadRequest {
		return fmt.Errorf("the request encoding is invalid; the request can't be parsed as a valid JSON")
	}

	if res.StatusCode == http.StatusUnauthorized {
		errStr := decodeJSONError(res)
		return fmt.Errorf("accessing a protected resource without authorization or with invalid credentials: \"%s\"", errStr)
	}

	if res.StatusCode == http.StatusPaymentRequired {
		return fmt.Errorf("the account associated with the API key making requests hits a quota that can be increased by upgrading the Airtable account plan")
	}

	if res.StatusCode == http.StatusForbidden {
		errStr := decodeJSONError(res)
		return fmt.Errorf("accessing a protected resource with API credentials that don't have access to that resource: \"%s\"", errStr)
	}

	if res.StatusCode == http.StatusNotFound {
		errStr := decodeJSONError(res)
		return fmt.Errorf("route or resource is not found. This error is returned when the request hits an undefined route, or if the resource doesn't exist (e.g. has been deleted): \"%s\"", errStr)
	}

	if res.StatusCode == http.StatusRequestEntityTooLarge {
		return fmt.Errorf("the request exceeded the maximum allowed payload size. You shouldn't encounter this under normal use")
	}

	if res.StatusCode == http.StatusUnprocessableEntity {
		errStr := decodeJSONError(res)
		return fmt.Errorf("the request data is invalid. This includes most of the base-specific validations. You will receive a detailed error message and code pointing to the exact issue, \"%s\"", errStr)
	}

	if res.StatusCode == http.StatusInternalServerError {
		return fmt.Errorf("the server encountered an unexpected condition")
	}

	if res.StatusCode == http.StatusBadGateway {
		return fmt.Errorf("airtable's servers are restarting or an unexpected outage is in progress. You should generally not receive this error, and requests are safe to retry")
	}

	if res.StatusCode == http.StatusServiceUnavailable {
		return fmt.Errorf("the server could not process your request in time. The server could be temporarily unavailable, or it could have timed out processing your request. You should retry the request with backoffs")
	}

	if method == DELETE {
		return nil
	}

	if response != nil {
		return json.NewDecoder(res.Body).Decode(&response)
	}

	return nil
}

// Attachment object may contain the following properties
type Attachment struct {
	ID         string `json:"id"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	URL        string `json:"url"`
	Filename   string `json:"filename"`
	Size       int    `json:"size"`
	Type       string `json:"type"`
	Thumbnails struct {
		Small struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"small"`
		Large struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"large"`
		Full struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"full"`
	} `json:"thumbnails"`
}

type AirtableItem struct {
	ID          string                 `json:"id"`
	CreatedTime time.Time              `json:"createdTime"`
	Fields      map[string]interface{} `json:"fields"`
}

type AirtableList struct {
	Records []AirtableItem `json:"records"`
	Offset  string         `json:"offset"`
}

var ErrEOL = errors.New("no more pages in list")

// A ListPager iterates the pages of a list by tracking
// page offsets.
type ListPager struct {
	at *Airtable
	pm Parameters
	lo AirtableList // "list object"

	done bool // track last page
}

func NewListPager(at *Airtable, pm Parameters) *ListPager {
	return &ListPager{at, pm, AirtableList{}, false}
}

// Next fetches the next available page from at.List(...) and
// returns the slice of records.
// If the list has been exhausted (no more pages)
// Next returns nil, ErrEOL.
func (p *ListPager) Next() (records []AirtableItem, err error) {
	if p.done {
		return nil, ErrEOL
	}

	// set params with last offset (if any), clear lo's offset
	p.pm.Offset, p.lo.Offset = p.lo.Offset, ""
	if err := p.at.List(p.pm, &p.lo); err != nil {
		return nil, err
	}

	p.done = p.lo.Offset == ""

	return p.lo.Records, nil
}

func (p *ListPager) Offset() string {
	return p.lo.Offset
}

// https://support.airtable.com/hc/en-us/articles/220340268-Supported-locale-modifiers-for-SET-LOCALE
type Local string

const (
	AF      Local = "af"
	ArMa    Local = "ar-ma"
	ArSa    Local = "ar-sa"
	ArTn    Local = "ar-tn"
	AR      Local = "ar"
	AZ      Local = "az"
	BE      Local = "be"
	BG      Local = "bg"
	BN      Local = "bn"
	BO      Local = "bo"
	BR      Local = "br"
	BS      Local = "bs"
	CA      Local = "ca"
	CS      Local = "cs"
	CV      Local = "cv"
	CY      Local = "cy"
	DA      Local = "da"
	DeAt    Local = "de-at"
	DE      Local = "de"
	EL      Local = "el"
	EnAu    Local = "en-au"
	EnCa    Local = "en-ca"
	EnGb    Local = "en-gb"
	EnIe    Local = "en-ie"
	EnNz    Local = "en-nz"
	EO      Local = "eo"
	ES      Local = "es"
	ET      Local = "et"
	EU      Local = "eu"
	FA      Local = "fa"
	FI      Local = "fi"
	FO      Local = "fo"
	FrCa    Local = "fr-ca"
	FrCh    Local = "fr-ch"
	FR      Local = "fr"
	FY      Local = "fy"
	GL      Local = "gl"
	HE      Local = "he"
	HI      Local = "hi"
	HR      Local = "hr"
	HU      Local = "hu"
	HyAm    Local = "hy-am"
	ID      Local = "id"
	IS      Local = "is"
	IT      Local = "it"
	JA      Local = "ja"
	JV      Local = "jv"
	KA      Local = "ka"
	KM      Local = "km"
	KO      Local = "ko"
	LB      Local = "lb"
	LT      Local = "lt"
	LV      Local = "lv"
	ME      Local = "me"
	MK      Local = "mk"
	ML      Local = "ml"
	MR      Local = "mr"
	MS      Local = "ms"
	MY      Local = "my"
	NB      Local = "nb"
	NE      Local = "ne"
	NL      Local = "nl"
	NN      Local = "nn"
	PL      Local = "pl"
	PtBr    Local = "pt-br"
	PT      Local = "pt"
	RO      Local = "ro"
	RU      Local = "ru"
	SI      Local = "si"
	SK      Local = "sk"
	SL      Local = "sl"
	SQ      Local = "sq"
	SrCyRL  Local = "sr-cyrl"
	SR      Local = "sr"
	SV      Local = "sv"
	TA      Local = "ta"
	TH      Local = "th"
	TlPh    Local = "tl-ph"
	TR      Local = "tr"
	TZL     Local = "tzl"
	TzMLaTN Local = "tzm-latn"
	TZM     Local = "tzm"
	UK      Local = "uk"
	UZ      Local = "uz"
	VI      Local = "vi"
	ZhCn    Local = "zh-cn"
	ZhTw    Local = "zh-tw"
)

// https://support.airtable.com/hc/en-us/articles/216141558-Supported-timezones-for-SET-TIMEZONE
type TimeZone string

const (
	AfricaAbidjan            TimeZone = "Africa/Abidjan"
	AfricaAccra              TimeZone = "Africa/Accra"
	AfricaAlgiers            TimeZone = "Africa/Algiers"
	AfricaBissau             TimeZone = "Africa/Bissau"
	AfricaCairo              TimeZone = "Africa/Cairo"
	AfricaCasablanca         TimeZone = "Africa/Casablanca"
	AfricaCeuta              TimeZone = "Africa/Ceuta"
	AfricaEl_Aaiun           TimeZone = "Africa/El_Aaiun"
	AfricaJohannesburg       TimeZone = "Africa/Johannesburg"
	AfricaKhartoum           TimeZone = "Africa/Khartoum"
	AfricaLagos              TimeZone = "Africa/Lagos"
	AfricaMaputo             TimeZone = "Africa/Maputo"
	AfricaMonrovia           TimeZone = "Africa/Monrovia"
	AfricaNairobi            TimeZone = "Africa/Nairobi"
	AfricaNdjamena           TimeZone = "Africa/Ndjamena"
	AfricaTripoli            TimeZone = "Africa/Tripoli"
	AfricaTunis              TimeZone = "Africa/Tunis"
	AfricaWindhoek           TimeZone = "Africa/Windhoek"
	AmericaAdak              TimeZone = "America/Adak"
	AmericaAnchorage         TimeZone = "America/Anchorage"
	AmericaAraguaina         TimeZone = "America/Araguaina"
	AmericaBuenos_Aires      TimeZone = "America/Argentina/Buenos_Aires"
	AmericaCatamarca         TimeZone = "America/Argentina/Catamarca"
	AmericaCordoba           TimeZone = "America/Argentina/Cordoba"
	AmericaJujuy             TimeZone = "America/Argentina/Jujuy"
	AmericaLa_Rioja          TimeZone = "America/Argentina/La_Rioja"
	AmericaMendoza           TimeZone = "America/Argentina/Mendoza"
	AmericaRio_Gallegos      TimeZone = "America/Argentina/Rio_Gallegos"
	AmericaSalta             TimeZone = "America/Argentina/Salta"
	AmericaSan_Juan          TimeZone = "America/Argentina/San_Juan"
	AmericaSan_Luis          TimeZone = "America/Argentina/San_Luis"
	AmericaTucuman           TimeZone = "America/Argentina/Tucuman"
	AmericaUshuaia           TimeZone = "America/Argentina/Ushuaia"
	AmericaAsuncion          TimeZone = "America/Asuncion"
	AmericaAtikokan          TimeZone = "America/Atikokan"
	AmericaBahia             TimeZone = "America/Bahia"
	AmericaBahia_Banderas    TimeZone = "America/Bahia_Banderas"
	AmericaBarbados          TimeZone = "America/Barbados"
	AmericaBelem             TimeZone = "America/Belem"
	AmericaBelize            TimeZone = "America/Belize"
	AmericaSablon            TimeZone = "America/Blanc-Sablon"
	AmericaBoa_Vista         TimeZone = "America/Boa_Vista"
	AmericaBogota            TimeZone = "America/Bogota"
	AmericaBoise             TimeZone = "America/Boise"
	AmericaCambridge_Bay     TimeZone = "America/Cambridge_Bay"
	AmericaCampo_Grande      TimeZone = "America/Campo_Grande"
	AmericaCancun            TimeZone = "America/Cancun"
	AmericaCaracas           TimeZone = "America/Caracas"
	AmericaCayenne           TimeZone = "America/Cayenne"
	AmericaChicago           TimeZone = "America/Chicago"
	AmericaChihuahua         TimeZone = "America/Chihuahua"
	AmericaCosta_Rica        TimeZone = "America/Costa_Rica"
	AmericaCreston           TimeZone = "America/Creston"
	AmericaCuiaba            TimeZone = "America/Cuiaba"
	AmericaCuracao           TimeZone = "America/Curacao"
	AmericaDanmarkshavn      TimeZone = "America/Danmarkshavn"
	AmericaDawson            TimeZone = "America/Dawson"
	AmericaDawson_Creek      TimeZone = "America/Dawson_Creek"
	AmericaDenver            TimeZone = "America/Denver"
	AmericaDetroit           TimeZone = "America/Detroit"
	AmericaEdmonton          TimeZone = "America/Edmonton"
	AmericaEirunepe          TimeZone = "America/Eirunepe"
	AmericaEl_Salvador       TimeZone = "America/El_Salvador"
	AmericaFort_Nelson       TimeZone = "America/Fort_Nelson"
	AmericaFortaleza         TimeZone = "America/Fortaleza"
	AmericaGlace_Bay         TimeZone = "America/Glace_Bay"
	AmericaGodthab           TimeZone = "America/Godthab"
	AmericaGoose_Bay         TimeZone = "America/Goose_Bay"
	AmericaGrand_Turk        TimeZone = "America/Grand_Turk"
	AmericaGuatemala         TimeZone = "America/Guatemala"
	AmericaGuayaquil         TimeZone = "America/Guayaquil"
	AmericaGuyana            TimeZone = "America/Guyana"
	AmericaHalifax           TimeZone = "America/Halifax"
	AmericaHavana            TimeZone = "America/Havana"
	AmericaHermosillo        TimeZone = "America/Hermosillo"
	AmericaIndianapolis      TimeZone = "America/Indiana/Indianapolis"
	AmericaKnox              TimeZone = "America/Indiana/Knox"
	AmericaMarengo           TimeZone = "America/Indiana/Marengo"
	AmericaPetersburg        TimeZone = "America/Indiana/Petersburg"
	AmericaTell_City         TimeZone = "America/Indiana/Tell_City"
	AmericaVevay             TimeZone = "America/Indiana/Vevay"
	AmericaVincennes         TimeZone = "America/Indiana/Vincennes"
	AmericaWinamac           TimeZone = "America/Indiana/Winamac"
	AmericaInuvik            TimeZone = "America/Inuvik"
	AmericaIqaluit           TimeZone = "America/Iqaluit"
	AmericaJamaica           TimeZone = "America/Jamaica"
	AmericaJuneau            TimeZone = "America/Juneau"
	AmericaLouisville        TimeZone = "America/Kentucky/Louisville"
	AmericaMonticello        TimeZone = "America/Kentucky/Monticello"
	AmericaLa_Paz            TimeZone = "America/La_Paz"
	AmericaLima              TimeZone = "America/Lima"
	AmericaLos_Angeles       TimeZone = "America/Los_Angeles"
	AmericaMaceio            TimeZone = "America/Maceio"
	AmericaManagua           TimeZone = "America/Managua"
	AmericaManaus            TimeZone = "America/Manaus"
	AmericaMartinique        TimeZone = "America/Martinique"
	AmericaMatamoros         TimeZone = "America/Matamoros"
	AmericaMazatlan          TimeZone = "America/Mazatlan"
	AmericaMenominee         TimeZone = "America/Menominee"
	AmericaMerida            TimeZone = "America/Merida"
	AmericaMetlakatla        TimeZone = "America/Metlakatla"
	AmericaMexico_City       TimeZone = "America/Mexico_City"
	AmericaMiquelon          TimeZone = "America/Miquelon"
	AmericaMoncton           TimeZone = "America/Moncton"
	AmericaMonterrey         TimeZone = "America/Monterrey"
	AmericaMontevideo        TimeZone = "America/Montevideo"
	AmericaNassau            TimeZone = "America/Nassau"
	AmericaNew_York          TimeZone = "America/New_York"
	AmericaNipigon           TimeZone = "America/Nipigon"
	AmericaNome              TimeZone = "America/Nome"
	AmericaNoronha           TimeZone = "America/Noronha"
	AmericaBeulah            TimeZone = "America/North_Dakota/Beulah"
	AmericaCenter            TimeZone = "America/North_Dakota/Center"
	AmericaNew_Salem         TimeZone = "America/North_Dakota/New_Salem"
	AmericaOjinaga           TimeZone = "America/Ojinaga"
	AmericaPanama            TimeZone = "America/Panama"
	AmericaPangnirtung       TimeZone = "America/Pangnirtung"
	AmericaParamaribo        TimeZone = "America/Paramaribo"
	AmericaPhoenix           TimeZone = "America/Phoenix"
	AmericaPrince            TimeZone = "America/Port-au-Prince"
	AmericaPort_of_Spain     TimeZone = "America/Port_of_Spain"
	AmericaPorto_Velho       TimeZone = "America/Porto_Velho"
	AmericaPuerto_Rico       TimeZone = "America/Puerto_Rico"
	AmericaRainy_River       TimeZone = "America/Rainy_River"
	AmericaRankin_Inlet      TimeZone = "America/Rankin_Inlet"
	AmericaRecife            TimeZone = "America/Recife"
	AmericaRegina            TimeZone = "America/Regina"
	AmericaResolute          TimeZone = "America/Resolute"
	AmericaRio_Branco        TimeZone = "America/Rio_Branco"
	AmericaSantarem          TimeZone = "America/Santarem"
	AmericaSantiago          TimeZone = "America/Santiago"
	AmericaSanto_Domingo     TimeZone = "America/Santo_Domingo"
	AmericaSao_Paulo         TimeZone = "America/Sao_Paulo"
	AmericaScoresbysund      TimeZone = "America/Scoresbysund"
	AmericaSitka             TimeZone = "America/Sitka"
	AmericaSt_Johns          TimeZone = "America/St_Johns"
	AmericaSwift_Current     TimeZone = "America/Swift_Current"
	AmericaTegucigalpa       TimeZone = "America/Tegucigalpa"
	AmericaThule             TimeZone = "America/Thule"
	AmericaThunder_Bay       TimeZone = "America/Thunder_Bay"
	AmericaTijuana           TimeZone = "America/Tijuana"
	AmericaToronto           TimeZone = "America/Toronto"
	AmericaVancouver         TimeZone = "America/Vancouver"
	AmericaWhitehorse        TimeZone = "America/Whitehorse"
	AmericaWinnipeg          TimeZone = "America/Winnipeg"
	AmericaYakutat           TimeZone = "America/Yakutat"
	AmericaYellowknife       TimeZone = "America/Yellowknife"
	AntarcticaCasey          TimeZone = "Antarctica/Casey"
	AntarcticaDavis          TimeZone = "Antarctica/Davis"
	AntarcticaDumontDUrville TimeZone = "Antarctica/DumontDUrville"
	AntarcticaMacquarie      TimeZone = "Antarctica/Macquarie"
	AntarcticaMawson         TimeZone = "Antarctica/Mawson"
	AntarcticaPalmer         TimeZone = "Antarctica/Palmer"
	AntarcticaRothera        TimeZone = "Antarctica/Rothera"
	AntarcticaSyowa          TimeZone = "Antarctica/Syowa"
	AntarcticaTroll          TimeZone = "Antarctica/Troll"
	AntarcticaVostok         TimeZone = "Antarctica/Vostok"
	AsiaAlmaty               TimeZone = "Asia/Almaty"
	AsiaAmman                TimeZone = "Asia/Amman"
	AsiaAnadyr               TimeZone = "Asia/Anadyr"
	AsiaAqtau                TimeZone = "Asia/Aqtau"
	AsiaAqtobe               TimeZone = "Asia/Aqtobe"
	AsiaAshgabat             TimeZone = "Asia/Ashgabat"
	AsiaBaghdad              TimeZone = "Asia/Baghdad"
	AsiaBaku                 TimeZone = "Asia/Baku"
	AsiaBangkok              TimeZone = "Asia/Bangkok"
	AsiaBarnaul              TimeZone = "Asia/Barnaul"
	AsiaBeirut               TimeZone = "Asia/Beirut"
	AsiaBishkek              TimeZone = "Asia/Bishkek"
	AsiaBrunei               TimeZone = "Asia/Brunei"
	AsiaChita                TimeZone = "Asia/Chita"
	AsiaChoibalsan           TimeZone = "Asia/Choibalsan"
	AsiaColombo              TimeZone = "Asia/Colombo"
	AsiaDamascus             TimeZone = "Asia/Damascus"
	AsiaDhaka                TimeZone = "Asia/Dhaka"
	AsiaDili                 TimeZone = "Asia/Dili"
	AsiaDubai                TimeZone = "Asia/Dubai"
	AsiaDushanbe             TimeZone = "Asia/Dushanbe"
	AsiaGaza                 TimeZone = "Asia/Gaza"
	AsiaHebron               TimeZone = "Asia/Hebron"
	AsiaHo_Chi_Minh          TimeZone = "Asia/Ho_Chi_Minh"
	AsiaHong_Kong            TimeZone = "Asia/Hong_Kong"
	AsiaHovd                 TimeZone = "Asia/Hovd"
	AsiaIrkutsk              TimeZone = "Asia/Irkutsk"
	AsiaJakarta              TimeZone = "Asia/Jakarta"
	AsiaJayapura             TimeZone = "Asia/Jayapura"
	AsiaJerusalem            TimeZone = "Asia/Jerusalem"
	AsiaKabul                TimeZone = "Asia/Kabul"
	AsiaKamchatka            TimeZone = "Asia/Kamchatka"
	AsiaKarachi              TimeZone = "Asia/Karachi"
	AsiaKathmandu            TimeZone = "Asia/Kathmandu"
	AsiaKhandyga             TimeZone = "Asia/Khandyga"
	AsiaKolkata              TimeZone = "Asia/Kolkata"
	AsiaKrasnoyarsk          TimeZone = "Asia/Krasnoyarsk"
	AsiaKuala_Lumpur         TimeZone = "Asia/Kuala_Lumpur"
	AsiaKuching              TimeZone = "Asia/Kuching"
	AsiaMacau                TimeZone = "Asia/Macau"
	AsiaMagadan              TimeZone = "Asia/Magadan"
	AsiaMakassar             TimeZone = "Asia/Makassar"
	AsiaManila               TimeZone = "Asia/Manila"
	AsiaNicosia              TimeZone = "Asia/Nicosia"
	AsiaNovokuznetsk         TimeZone = "Asia/Novokuznetsk"
	AsiaNovosibirsk          TimeZone = "Asia/Novosibirsk"
	AsiaOmsk                 TimeZone = "Asia/Omsk"
	AsiaOral                 TimeZone = "Asia/Oral"
	AsiaPontianak            TimeZone = "Asia/Pontianak"
	AsiaPyongyang            TimeZone = "Asia/Pyongyang"
	AsiaQatar                TimeZone = "Asia/Qatar"
	AsiaQyzylorda            TimeZone = "Asia/Qyzylorda"
	AsiaRangoon              TimeZone = "Asia/Rangoon"
	AsiaRiyadh               TimeZone = "Asia/Riyadh"
	AsiaSakhalin             TimeZone = "Asia/Sakhalin"
	AsiaSamarkand            TimeZone = "Asia/Samarkand"
	AsiaSeoul                TimeZone = "Asia/Seoul"
	AsiaShanghai             TimeZone = "Asia/Shanghai"
	AsiaSingapore            TimeZone = "Asia/Singapore"
	AsiaSrednekolymsk        TimeZone = "Asia/Srednekolymsk"
	AsiaTaipei               TimeZone = "Asia/Taipei"
	AsiaTashkent             TimeZone = "Asia/Tashkent"
	AsiaTbilisi              TimeZone = "Asia/Tbilisi"
	AsiaTehran               TimeZone = "Asia/Tehran"
	AsiaThimphu              TimeZone = "Asia/Thimphu"
	AsiaTokyo                TimeZone = "Asia/Tokyo"
	AsiaTomsk                TimeZone = "Asia/Tomsk"
	AsiaUlaanbaatar          TimeZone = "Asia/Ulaanbaatar"
	AsiaUrumqi               TimeZone = "Asia/Urumqi"
	AsiaNera                 TimeZone = "Asia/Ust-Nera"
	AsiaVladivostok          TimeZone = "Asia/Vladivostok"
	AsiaYakutsk              TimeZone = "Asia/Yakutsk"
	AsiaYekaterinburg        TimeZone = "Asia/Yekaterinburg"
	AsiaYerevan              TimeZone = "Asia/Yerevan"
	AtlanticAzores           TimeZone = "Atlantic/Azores"
	AtlanticBermuda          TimeZone = "Atlantic/Bermuda"
	AtlanticCanary           TimeZone = "Atlantic/Canary"
	AtlanticCape_Verde       TimeZone = "Atlantic/Cape_Verde"
	AtlanticFaroe            TimeZone = "Atlantic/Faroe"
	AtlanticMadeira          TimeZone = "Atlantic/Madeira"
	AtlanticReykjavik        TimeZone = "Atlantic/Reykjavik"
	AtlanticSouth_Georgia    TimeZone = "Atlantic/South_Georgia"
	AtlanticStanley          TimeZone = "Atlantic/Stanley"
	AustraliaAdelaide        TimeZone = "Australia/Adelaide"
	AustraliaBrisbane        TimeZone = "Australia/Brisbane"
	AustraliaBroken_Hill     TimeZone = "Australia/Broken_Hill"
	AustraliaCurrie          TimeZone = "Australia/Currie"
	AustraliaDarwin          TimeZone = "Australia/Darwin"
	AustraliaEucla           TimeZone = "Australia/Eucla"
	AustraliaHobart          TimeZone = "Australia/Hobart"
	AustraliaLindeman        TimeZone = "Australia/Lindeman"
	AustraliaLord_Howe       TimeZone = "Australia/Lord_Howe"
	AustraliaMelbourne       TimeZone = "Australia/Melbourne"
	AustraliaPerth           TimeZone = "Australia/Perth"
	AustraliaSydney          TimeZone = "Australia/Sydney"
	GMT                      TimeZone = "GMT"
	EuropeAmsterdam          TimeZone = "Europe/Amsterdam"
	EuropeAndorra            TimeZone = "Europe/Andorra"
	EuropeAstrakhan          TimeZone = "Europe/Astrakhan"
	EuropeAthens             TimeZone = "Europe/Athens"
	EuropeBelgrade           TimeZone = "Europe/Belgrade"
	EuropeBerlin             TimeZone = "Europe/Berlin"
	EuropeBrussels           TimeZone = "Europe/Brussels"
	EuropeBucharest          TimeZone = "Europe/Bucharest"
	EuropeBudapest           TimeZone = "Europe/Budapest"
	EuropeChisinau           TimeZone = "Europe/Chisinau"
	EuropeCopenhagen         TimeZone = "Europe/Copenhagen"
	EuropeDublin             TimeZone = "Europe/Dublin"
	EuropeGibraltar          TimeZone = "Europe/Gibraltar"
	EuropeHelsinki           TimeZone = "Europe/Helsinki"
	EuropeIstanbul           TimeZone = "Europe/Istanbul"
	EuropeKaliningrad        TimeZone = "Europe/Kaliningrad"
	EuropeKiev               TimeZone = "Europe/Kiev"
	EuropeKirov              TimeZone = "Europe/Kirov"
	EuropeLisbon             TimeZone = "Europe/Lisbon"
	EuropeLondon             TimeZone = "Europe/London"
	EuropeLuxembourg         TimeZone = "Europe/Luxembourg"
	EuropeMadrid             TimeZone = "Europe/Madrid"
	EuropeMalta              TimeZone = "Europe/Malta"
	EuropeMinsk              TimeZone = "Europe/Minsk"
	EuropeMonaco             TimeZone = "Europe/Monaco"
	EuropeMoscow             TimeZone = "Europe/Moscow"
	EuropeOslo               TimeZone = "Europe/Oslo"
	EuropeParis              TimeZone = "Europe/Paris"
	EuropePrague             TimeZone = "Europe/Prague"
	EuropeRiga               TimeZone = "Europe/Riga"
	EuropeRome               TimeZone = "Europe/Rome"
	EuropeSamara             TimeZone = "Europe/Samara"
	EuropeSimferopol         TimeZone = "Europe/Simferopol"
	EuropeSofia              TimeZone = "Europe/Sofia"
	EuropeStockholm          TimeZone = "Europe/Stockholm"
	EuropeTallinn            TimeZone = "Europe/Tallinn"
	EuropeTirane             TimeZone = "Europe/Tirane"
	EuropeUlyanovsk          TimeZone = "Europe/Ulyanovsk"
	EuropeUzhgorod           TimeZone = "Europe/Uzhgorod"
	EuropeVienna             TimeZone = "Europe/Vienna"
	EuropeVilnius            TimeZone = "Europe/Vilnius"
	EuropeVolgograd          TimeZone = "Europe/Volgograd"
	EuropeWarsaw             TimeZone = "Europe/Warsaw"
	EuropeZaporozhye         TimeZone = "Europe/Zaporozhye"
	EuropeZurich             TimeZone = "Europe/Zurich"
	IndianChagos             TimeZone = "Indian/Chagos"
	IndianChristmas          TimeZone = "Indian/Christmas"
	IndianCocos              TimeZone = "Indian/Cocos"
	IndianKerguelen          TimeZone = "Indian/Kerguelen"
	IndianMahe               TimeZone = "Indian/Mahe"
	IndianMaldives           TimeZone = "Indian/Maldives"
	IndianMauritius          TimeZone = "Indian/Mauritius"
	IndianReunion            TimeZone = "Indian/Reunion"
	PacificApia              TimeZone = "Pacific/Apia"
	PacificAuckland          TimeZone = "Pacific/Auckland"
	PacificBougainville      TimeZone = "Pacific/Bougainville"
	PacificChatham           TimeZone = "Pacific/Chatham"
	PacificChuuk             TimeZone = "Pacific/Chuuk"
	PacificEaster            TimeZone = "Pacific/Easter"
	PacificEfate             TimeZone = "Pacific/Efate"
	PacificEnderbury         TimeZone = "Pacific/Enderbury"
	PacificFakaofo           TimeZone = "Pacific/Fakaofo"
	PacificFiji              TimeZone = "Pacific/Fiji"
	PacificFunafuti          TimeZone = "Pacific/Funafuti"
	PacificGalapagos         TimeZone = "Pacific/Galapagos"
	PacificGambier           TimeZone = "Pacific/Gambier"
	PacificGuadalcanal       TimeZone = "Pacific/Guadalcanal"
	PacificGuam              TimeZone = "Pacific/Guam"
	PacificHonolulu          TimeZone = "Pacific/Honolulu"
	PacificKiritimati        TimeZone = "Pacific/Kiritimati"
	PacificKosrae            TimeZone = "Pacific/Kosrae"
	PacificKwajalein         TimeZone = "Pacific/Kwajalein"
	PacificMajuro            TimeZone = "Pacific/Majuro"
	PacificMarquesas         TimeZone = "Pacific/Marquesas"
	PacificNauru             TimeZone = "Pacific/Nauru"
	PacificNiue              TimeZone = "Pacific/Niue"
	PacificNorfolk           TimeZone = "Pacific/Norfolk"
	PacificNoumea            TimeZone = "Pacific/Noumea"
	PacificPago_Pago         TimeZone = "Pacific/Pago_Pago"
	PacificPalau             TimeZone = "Pacific/Palau"
	PacificPitcairn          TimeZone = "Pacific/Pitcairn"
	PacificPohnpei           TimeZone = "Pacific/Pohnpei"
	PacificPort_Moresby      TimeZone = "Pacific/Port_Moresby"
	PacificRarotonga         TimeZone = "Pacific/Rarotonga"
	PacificTahiti            TimeZone = "Pacific/Tahiti"
	PacificTarawa            TimeZone = "Pacific/Tarawa"
	PacificTongatapu         TimeZone = "Pacific/Tongatapu"
	PacificWake              TimeZone = "Pacific/Wake"
	PacificWallis            TimeZone = "Pacific/Wallis"
)

type SortDirection string

const (
	Ascending  SortDirection = "asc"
	Descending SortDirection = "desc"
)
