package structures

type SonarrSettings struct {
	ID           int     `json:"id"`
	APIKey       *string `json:"api_key"`
	URL          *string `json:"url"`
	Language     *string `json:"language"`
	Quality      *int    `json:"quality"`
	RootFolder   *int    `json:"root_folder"`
	SeasonFolder *bool   `json:"season_folder"`
}

type SonarrQualityProfile struct {
	Name              string              `json:"name"`
	UpgradeAllowed    bool                `json:"upgradeAllowed"`
	Cutoff            int                 `json:"cutoff"`
	Items             []SonarrQualityItem `json:"items"`
	MinFormatScore    int                 `json:"minFormatScore"`
	CutoffFormatScore int                 `json:"cutoffFormatScore"`
	FormatItems       []SonarrQualityItem `json:"formatItems"`
	ID                int                 `json:"id"`
}

type SonarrQualityItem struct {
	Quality *SonarrQuality      `json:"quality,omitempty"` // Optional field for nested items
	Items   []SonarrQualityItem `json:"items"`
	Allowed bool                `json:"allowed"`
	Name    string              `json:"name,omitempty"` // Optional field for named items
	ID      int                 `json:"id,omitempty"`   // Optional field for named items
}

type SonarrQuality struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Source     string `json:"source"`
	Resolution int    `json:"resolution"`
}

type SonarrRootFolder struct {
	ID         int    `json:"id"`
	Path       string `json:"path"`
	Accessible bool   `json:"accessible"`
	FreeSpace  int64  `json:"free_space"`
}
