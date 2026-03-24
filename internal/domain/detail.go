package domain

type DetailEpisode struct {
	Name string
	URL  string
}

type DetailPlayer struct {
	Name     string
	Episodes []DetailEpisode
}

type DetailItem struct {
	VodName        string
	VodSub         string
	TypeName       string
	VodTag         string
	VodClass       string
	VodActor       string
	VodDirector    string
	VodArea        string
	VodLang        string
	VodYear        string
	VodHits        int
	VodScore       string
	VodDoubanScore string
	VodTime        string
	VodPubDate     string
	VodTotal       int
	VodContent     string
	VodRemarks     string
	Players        []DetailPlayer
}

type DetailResult struct {
	SeriesIndex int
	DomainIndex int
	DomainURL   string
	Item        *DetailItem
	Err         error
}
