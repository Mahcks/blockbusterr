export interface SonarrSettings {
    id: number;
    api_key?: string;
    url?: string;
    language?: string;
    quality?: number;
    root_folder?: number;
    season_folder?: boolean;
}

export interface SonarrRootFolder {
    id: number;
    path: string;
    accessible: boolean;
    free_space: number;
}

export interface SonarrQualityProfile {
    name: string;
    upgradeAllowed: boolean;
    cutoff: number;
    items: SonarrQualityItem[];
    minFormatScore: number;
    cutoffFormatScore: number;
    formatItems: SonarrQualityItem[];
    id: number;
};

export interface SonarrQualityItem {
    quality?: SonarrQuality;  // Optional field for nested items
    items: SonarrQualityItem[];
    allowed: boolean;
    name?: string;  // Optional field for named items
    id?: number;    // Optional field for named items
};

export interface SonarrQuality {
    id: number;
    name: string;
    source: string;
    resolution: number;
};
