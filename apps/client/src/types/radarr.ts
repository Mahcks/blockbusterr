export interface RadarrSettings {
    id: number;
    api_key?: string;
    base_url?: string;
    minimum_availability?: string;
    quality?: number;
    root_folder?: number;
}

export interface RadarrRootFolder {
    id: number;
    path: string;
    accessible: boolean;
    free_space: number;
    unmapped_folders: RadarrRootFolderUnmappedFolder[];
}

export interface RadarrRootFolderUnmappedFolder {
    name: string;
    path: string;
    relative_path: string;
}

export interface RadarrQualityProfile {
    id: number;
    name: string;
    upgradeAllowed: boolean;
    cutoff: number;
    minFormatScore: number;
    cutoffFormatScore: number;
    items: RadarrQualityProfileItem[];
    language: RadarrQualityProfileLanguage;
}

export interface RadarrQualityProfileItem {
    quality: RadarrQuality;
    allowed: boolean;
}

export interface RadarrQuality {
    id: number;
    name: string;
    source: string;
    resolution: number;
    modifier: string;
}

export interface RadarrQualityProfileLanguage {
    id: number;
    name: string;
}
