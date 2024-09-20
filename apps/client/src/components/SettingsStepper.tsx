import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn, formatBytes } from "@/lib/utils";
import { Separator } from "@/components/ui/separator";
import { APIErrorBody } from "@/types/api_error";
import { useNavigate } from "react-router-dom";
import { useSetupStatus } from "@/context/SetupContext";

type Step = {
  title: string;
  description: string;
};

type Mode = {
  value: string;
  label: string;
};

type Folder = {
  id: number;
  path: string;
  free_space: number;
};

type QualityProfile = {
  id: number;
  name: string;
};

const steps: Step[] = [
  {
    title: "Trakt Details",
    description: "Enter your Trakt client ID and client secret",
  },
  {
    title: "OMDb API",
    description: "Enter your OMDb API key",
  },
  { title: "Mode Selection", description: "Choose your preferred mode" },
  {
    title: "Additional Settings",
    description: "Configure additional settings based on your mode selection",
  },
];

const modes: Mode[] = [
  { value: "ombi", label: "Ombi" },
  { value: "radarr-sonarr", label: "Radarr/Sonarr" },
];

export default function SettingsStepper() {
  const navigate = useNavigate();
  const { checkSetupStatus } = useSetupStatus();

  const [currentStep, setCurrentStep] = useState(0);
  const [formData, setFormData] = useState({
    traktClientId: "",
    traktClientSecret: "",
    omdbApiKey: "",
    selectedMode: modes[0].value,
    settings: {} as Record<string, string>,
  });
  const [sonarrRootFolders, setSonarrRootFolders] = useState<Folder[]>([]);
  const [sonarrQualityProfiles, setSonarrQualityProfiles] = useState<
    QualityProfile[]
  >([]);
  const [radarrRootFolders, setRadarrRootFolders] = useState<Folder[]>([]);
  const [radarrQualityProfiles, setRadarrQualityProfiles] = useState<
    QualityProfile[]
  >([]);
  const [sonarrError, setSonarrError] = useState<string | null>(null);
  const [radarrError, setRadarrError] = useState<string | null>(null);

  const isFormValid = () => {
    const {
      traktClientId,
      traktClientSecret,
      omdbApiKey,
      selectedMode,
      settings,
    } = formData;

    if (selectedMode === "ombi") {
      return (
        traktClientId &&
        traktClientSecret &&
        omdbApiKey &&
        settings["ombi-base-url"] &&
        settings["ombi-api-key"]
      );
    } else if (selectedMode === "radarr-sonarr") {
      return (
        traktClientId &&
        traktClientSecret &&
        omdbApiKey &&
        settings["sonarr-base-url"] &&
        settings["sonarr-api-key"] &&
        sonarrRootFolders.length > 0 &&
        sonarrQualityProfiles.length > 0 &&
        settings["sonarr-root-folder"] &&
        settings["sonarr-quality-profile"] &&
        settings["radarr-base-url"] &&
        settings["radarr-api-key"] &&
        radarrRootFolders.length > 0 &&
        radarrQualityProfiles.length > 0 &&
        settings["radarr-root-folder"] &&
        settings["radarr-quality-profile"]
      );
    }
    return false;
  };

  const handleAPIError = async (response: Response) => {
    const contentType = response.headers.get("content-type");

    if (contentType && contentType.includes("application/json")) {
      try {
        const errorBody: APIErrorBody = await response.json();
        return errorBody.error.error;
      } catch (err) {
        return `Error parsing JSON: ${err}`;
      }
    }

    return `Error ${response.status}: ${response.statusText}`;
  };

  const handleNext = async () => {
    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1);
    } else if (isFormValid()) {
      const allSettings = {
        traktClientId: formData.traktClientId,
        traktClientSecret: formData.traktClientSecret,
        omdbApiKey: formData.omdbApiKey,
        selectedMode: formData.selectedMode,
        ...formData.settings,
      };

      console.log("Submitting settings", allSettings);

      try {
        // First POST request to /settings/setup
        const response = await fetch(
          `${import.meta.env.VITE_API_URL}/settings/setup`,
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify(allSettings),
          }
        );

        if (!response.ok) {
          const errorMessage = await handleAPIError(response);
          throw new Error(errorMessage);
        }

        // Handle successful submission
        console.log("Settings saved successfully");

        // Second POST request to /settings?key=SETUP_COMPLETE
        const setupCompleteResponse = await fetch(
          `${import.meta.env.VITE_API_URL}/settings?key=SETUP_COMPLETE`,
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify({
              key: "SETUP_COMPLETE",
              value: "true",
              type: "boolean",
            }),
          }
        );

        if (!setupCompleteResponse.ok) {
          const errorMessage = await handleAPIError(setupCompleteResponse);
          throw new Error(errorMessage);
        }

        // Handle successful setting of SETUP_COMPLETE
        console.log("Setup complete flag set successfully");
        await checkSetupStatus(); // Update the setup status in the context

        // Redirect the user to the dashboard
        navigate("/");
      } catch (error) {
        console.error(
          error instanceof Error ? error.message : "An error occurred"
        );
      }
    }
  };

  const handlePrevious = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  const handleInputChange = (field: string, value: string) => {
    setFormData((prevData) => ({ ...prevData, [field]: value }));
  };

  const handleSettingChange = (key: string, value: string) => {
    setFormData((prevData) => {
      if (prevData.settings[key] !== value) {
        return {
          ...prevData,
          settings: { ...prevData.settings, [key]: value },
        };
      } else {
        return prevData;
      }
    });
  };

  // Event handler for saving Sonarr settings
  const handleSaveSonarrSettings = async () => {
    const { settings } = formData;
    const baseUrl = settings["sonarr-base-url"];
    const apiKey = settings["sonarr-api-key"];

    if (!baseUrl || !apiKey) {
      setSonarrError("Please provide both the base URL and API key.");
      return;
    }

    try {
      const rootFoldersResponse = await fetch(
        `${import.meta.env.VITE_API_URL}/sonarr/rootfolders?url=${baseUrl}`,
        {
          headers: {
            "X-Api-Key": apiKey,
          },
        }
      );
      if (!rootFoldersResponse.ok) {
        const errorMessage = await handleAPIError(rootFoldersResponse);
        throw new Error(errorMessage);
      }
      const rootFolders = await rootFoldersResponse.json();
      setSonarrRootFolders(rootFolders);

      const qualityProfilesResponse = await fetch(
        `${import.meta.env.VITE_API_URL}/sonarr/profiles?url=${baseUrl}`,
        {
          headers: {
            "X-Api-Key": apiKey,
          },
        }
      );
      if (!qualityProfilesResponse.ok) {
        const errorMessage = await handleAPIError(qualityProfilesResponse);
        throw new Error(errorMessage);
      }
      const qualityProfiles = await qualityProfilesResponse.json();
      setSonarrQualityProfiles(qualityProfiles);
      setSonarrError(null);
    } catch (error: unknown) {
      setSonarrError(
        error instanceof Error ? error.message : "An error occurred"
      );
      setSonarrRootFolders([]);
      setSonarrQualityProfiles([]);
    }
  };

  // Event handler for saving Radarr settings
  const handleSaveRadarrSettings = async () => {
    const { settings } = formData;
    const baseUrl = settings["radarr-base-url"];
    const apiKey = settings["radarr-api-key"];

    if (!baseUrl || !apiKey) {
      setRadarrError("Please provide both the base URL and API key.");
      return;
    }

    try {
      const rootFoldersResponse = await fetch(
        `${import.meta.env.VITE_API_URL}/radarr/rootfolders?url=${baseUrl}`,
        {
          headers: {
            "X-Api-Key": apiKey,
          },
        }
      );
      if (!rootFoldersResponse.ok) {
        const errorMessage = await handleAPIError(rootFoldersResponse);
        throw new Error(errorMessage);
      }
      const rootFolders = await rootFoldersResponse.json();
      setRadarrRootFolders(rootFolders);

      const qualityProfilesResponse = await fetch(
        `${import.meta.env.VITE_API_URL}/radarr/profiles?url=${baseUrl}`,
        {
          headers: {
            "X-Api-Key": apiKey,
          },
        }
      );
      if (!qualityProfilesResponse.ok) {
        const errorMessage = await handleAPIError(qualityProfilesResponse);
        throw new Error(errorMessage);
      }
      const qualityProfiles = await qualityProfilesResponse.json();
      setRadarrQualityProfiles(qualityProfiles);
      setRadarrError(null);
    } catch (error: unknown) {
      setRadarrError(
        error instanceof Error ? error.message : "An error occurred"
      );
      setRadarrRootFolders([]);
      setRadarrQualityProfiles([]);
    }
  };

  const renderStepContent = () => {
    switch (currentStep) {
      case 0:
        return (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="trakt-client-id">Trakt Client ID</Label>
              <Input
                id="trakt-client-id"
                value={formData.traktClientId}
                onChange={(e) =>
                  handleInputChange("traktClientId", e.target.value)
                }
                placeholder="Enter your Trakt Client ID"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="trakt-client-secret">Trakt Client Secret</Label>
              <Input
                id="trakt-client-secret"
                type="password"
                value={formData.traktClientSecret}
                onChange={(e) =>
                  handleInputChange("traktClientSecret", e.target.value)
                }
                placeholder="Enter your Trakt Client Secret"
              />
            </div>
          </div>
        );

      case 1:
        return (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="omdb-api-key">OMDb API Key</Label>
              <Input
                id="omdb-api-key"
                value={formData.omdbApiKey}
                onChange={(e) =>
                  handleInputChange("omdbApiKey", e.target.value)
                }
                placeholder="API key here..."
              />
            </div>
          </div>
        );

      case 2:
        return (
          <RadioGroup
            value={formData.selectedMode}
            onValueChange={(value) => handleInputChange("selectedMode", value)}
          >
            {modes.map((mode) => (
              <div key={mode.value} className="flex items-center space-x-2">
                <RadioGroupItem value={mode.value} id={mode.value} />
                <Label htmlFor={mode.value}>{mode.label}</Label>
              </div>
            ))}
          </RadioGroup>
        );
      case 3:
        return renderModeSettings();
      default:
        return null;
    }
  };

  const renderModeSettings = () => {
    switch (formData.selectedMode) {
      case "ombi":
        return (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="ombi-base-url">Ombi Base URL</Label>
              <Input
                id="ombi-base-url"
                value={formData.settings["ombi-base-url"] || ""}
                onChange={(e) =>
                  handleSettingChange("ombi-base-url", e.target.value)
                }
                placeholder="Enter Ombi Base URL"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="ombi-api-key">Ombi API Key</Label>
              <Input
                id="ombi-api-key"
                type="password"
                value={formData.settings["ombi-api-key"] || ""}
                onChange={(e) =>
                  handleSettingChange("ombi-api-key", e.target.value)
                }
                placeholder="Enter Ombi API Key"
              />
            </div>
          </div>
        );
      case "radarr-sonarr": {
        return (
          <div className="space-y-6">
            {/* Sonarr Settings */}
            <div className="space-y-4">
              <h2 className="text-lg font-medium">Sonarr Settings</h2>
              {sonarrError && (
                <div className="text-red-500">Error: {sonarrError}</div>
              )}
              <div className="grid grid-cols-2 gap-4 items-end">
                <div className="space-y-2">
                  <Label htmlFor="sonarr-base-url">Sonarr Base URL</Label>
                  <Input
                    id="sonarr-base-url"
                    value={formData.settings["sonarr-base-url"] || ""}
                    onChange={(e) =>
                      handleSettingChange("sonarr-base-url", e.target.value)
                    }
                    placeholder="http://localhost:8989"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="sonarr-api-key">Sonarr API Key</Label>
                  <Input
                    id="sonarr-api-key"
                    type="password"
                    value={formData.settings["sonarr-api-key"] || ""}
                    onChange={(e) =>
                      handleSettingChange("sonarr-api-key", e.target.value)
                    }
                    placeholder="Enter Sonarr API Key"
                  />
                </div>
                <div>
                  <Button onClick={handleSaveSonarrSettings}>
                    Save Sonarr Settings
                  </Button>
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="sonarr-root-folder">Sonarr Root Folder</Label>
                  <Select
                    onValueChange={(value) =>
                      handleSettingChange("sonarr-root-folder", value)
                    }
                    value={formData.settings["sonarr-root-folder"] || ""}
                    disabled={!sonarrRootFolders.length}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select a root folder to use..." />
                    </SelectTrigger>
                    <SelectContent>
                      {sonarrRootFolders.length > 0 ? (
                        sonarrRootFolders.map((folder) => (
                          <SelectItem
                            key={folder.id}
                            value={folder.id.toString()}
                          >
                            {folder.path} (Free space:{" "}
                            {formatBytes(folder.free_space)})
                          </SelectItem>
                        ))
                      ) : (
                        <SelectItem value="no-folder">
                          No folders available
                        </SelectItem>
                      )}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="sonarr-quality-profile">
                    Sonarr Quality Profile
                  </Label>
                  <Select
                    onValueChange={(value) =>
                      handleSettingChange("sonarr-quality-profile", value)
                    }
                    value={formData.settings["sonarr-quality-profile"] || ""}
                    disabled={!sonarrQualityProfiles.length}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select a quality profile..." />
                    </SelectTrigger>
                    <SelectContent>
                      {sonarrQualityProfiles.length > 0 ? (
                        sonarrQualityProfiles.map((profile) => (
                          <SelectItem
                            key={profile.id}
                            value={profile.id.toString()}
                          >
                            {profile.name}
                          </SelectItem>
                        ))
                      ) : (
                        <SelectItem value="no-profile">
                          No profiles available
                        </SelectItem>
                      )}
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </div>

            <Separator />

            {/* Radarr Settings */}
            <div className="space-y-4">
              <h2 className="text-lg font-medium">Radarr Settings</h2>
              {radarrError && (
                <div className="text-red-500">Error: {radarrError}</div>
              )}
              <div className="grid grid-cols-2 gap-4 items-end">
                <div className="space-y-2">
                  <Label htmlFor="radarr-base-url">Radarr Base URL</Label>
                  <Input
                    id="radarr-base-url"
                    value={formData.settings["radarr-base-url"] || ""}
                    onChange={(e) =>
                      handleSettingChange("radarr-base-url", e.target.value)
                    }
                    placeholder="http://localhost:7878"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="radarr-api-key">Radarr API Key</Label>
                  <Input
                    id="radarr-api-key"
                    type="password"
                    value={formData.settings["radarr-api-key"] || ""}
                    onChange={(e) =>
                      handleSettingChange("radarr-api-key", e.target.value)
                    }
                    placeholder="Enter Radarr API Key"
                  />
                </div>
                <div>
                  <Button onClick={handleSaveRadarrSettings}>
                    Save Radarr Settings
                  </Button>
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="radarr-root-folder">Radarr Root Folder</Label>
                  <Select
                    onValueChange={(value) =>
                      handleSettingChange("radarr-root-folder", value)
                    }
                    value={formData.settings["radarr-root-folder"] || ""}
                    disabled={!radarrRootFolders.length}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select a root folder to use..." />
                    </SelectTrigger>
                    <SelectContent>
                      {radarrRootFolders.length > 0 ? (
                        radarrRootFolders.map((folder) => (
                          <SelectItem
                            key={folder.id}
                            value={folder.id.toString()}
                          >
                            {folder.path} (Free space:{" "}
                            {formatBytes(folder.free_space)})
                          </SelectItem>
                        ))
                      ) : (
                        <SelectItem value="no-folder">
                          No folders available
                        </SelectItem>
                      )}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="radarr-quality-profile">
                    Radarr Quality Profile
                  </Label>
                  <Select
                    onValueChange={(value) =>
                      handleSettingChange("radarr-quality-profile", value)
                    }
                    value={formData.settings["radarr-quality-profile"] || ""}
                    disabled={!radarrQualityProfiles.length}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select a quality profile..." />
                    </SelectTrigger>
                    <SelectContent>
                      {radarrQualityProfiles.length > 0 ? (
                        radarrQualityProfiles.map((profile) => (
                          <SelectItem
                            key={profile.id}
                            value={profile.id.toString()}
                          >
                            {profile.name}
                          </SelectItem>
                        ))
                      ) : (
                        <SelectItem value="no-profile">
                          No profiles available
                        </SelectItem>
                      )}
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </div>
          </div>
        );
      }
      default:
        return null;
    }
  };

  return (
    <Card className="w-full max-w-[550px] mx-auto my-8">
      <CardHeader>
        <CardTitle>{steps[currentStep].title}</CardTitle>
        <CardDescription>{steps[currentStep].description}</CardDescription>
      </CardHeader>
      <CardContent>{renderStepContent()}</CardContent>
      <CardFooter className="flex justify-between">
        <Button onClick={handlePrevious} disabled={currentStep === 0}>
          Previous
        </Button>
        <div className="flex space-x-2">
          {steps.map((_, index) => (
            <div
              key={index}
              className={cn(
                "h-2 w-2 rounded-full",
                index === currentStep ? "bg-primary" : "bg-muted"
              )}
            />
          ))}
        </div>
        <Button
          onClick={handleNext}
          disabled={currentStep === steps.length - 1 && !isFormValid()}
        >
          {currentStep === steps.length - 1 ? "Finish" : "Next"}
        </Button>
      </CardFooter>
    </Card>
  );
}
