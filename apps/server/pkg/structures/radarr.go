package structures

type RadarrSettings struct {
	ID                  int     `json:"id"`
	APIKey              *string `json:"api_key"`
	URL                 *string `json:"base_url"`
	MinimumAvailability *string `json:"minimum_availability"`
	Quality             *int    `json:"quality"`
	RootFolder          *int    `json:"root_folder"`
}

type RadarrQualityProfile struct {
	ID                int                          `json:"id"`
	Name              string                       `json:"name"`
	UpgradeAllowed    bool                         `json:"upgrade_allowed"`
	Cutoff            int                          `json:"cutoff"`
	MinFormatScore    int                          `json:"min_format_score"`
	CutoffFormatScore int                          `json:"cutoff_format_score"`
	Items             RadarrQualityProfileItem     `json:"items"`
	Language          RadarrQualityProfileLanguage `json:"language"`
}

type RadarrQualityProfileItem []RadarrQualityProfileItemData

type RadarrQualityProfileItemData struct {
	Quality RadarrQuality `json:"quality"`
	Allowed bool          `json:"allowed"`
}

type RadarrQuality struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Source     string `json:"source"`
	Resolution int    `json:"resolution"`
	Modifier   string `json:"modifier"`
}

type RadarrQualityProfileLanguage struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type RadarrRootFolder struct {
	ID              int                              `json:"id"`
	Path            string                           `json:"path"`
	Accessible      bool                             `json:"accessible"`
	FreeSpace       int64                            `json:"free_space"`
	UnmappedFolders []RadarrRootFolderUnmappedFolder `json:"unmapped_folders"`
}

type RadarrRootFolderUnmappedFolder struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	RelativePath string `json:"relative_path"`
}
