"use client"

import { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { cn, formatBytes } from "@/lib/utils"
import { Separator } from "@/components/ui/separator"

type Step = {
  title: string
  description: string
}

type Mode = {
  value: string
  label: string
}

type Folder = {
  id: number
  path: string
  free_space: number
}

type QualityProfile = {
  id: number
  name: string
}

const steps: Step[] = [
  {
    title: "Trakt Details",
    description: "Enter your Trakt client ID and client secret",
  },
  {
    title: "Mode Selection",
    description: "Choose your preferred mode",
  },
  {
    title: "Additional Settings",
    description: "Configure additional settings based on your mode selection",
  },
]

const modes: Mode[] = [
  { value: "ombi", label: "Ombi" },
  { value: "radarr-sonarr", label: "Radarr/Sonarr" },
]

export default function SettingsStepper() {
  const [currentStep, setCurrentStep] = useState(0)
  const [traktClientId, setTraktClientId] = useState("")
  const [traktClientSecret, setTraktClientSecret] = useState("")
  const [selectedMode, setSelectedMode] = useState<string>(modes[0].value)
  const [settings, setSettings] = useState<Record<string, string>>({})
  const [sonarrRootFolders, setSonarrRootFolders] = useState<Folder[]>([])
  const [sonarrQualityProfiles, setSonarrQualityProfiles] = useState<QualityProfile[]>([])
  const [radarrRootFolders, setRadarrRootFolders] = useState<Folder[]>([])
  const [radarrQualityProfiles, setRadarrQualityProfiles] = useState<QualityProfile[]>([])

  useEffect(() => {
    const fetchSonarrData = async () => {
      try {
        const rootFoldersResponse = await fetch(`${import.meta.env.VITE_API_URL}/sonarr/rootfolders`);
        const rootFolders = await rootFoldersResponse.json();
        setSonarrRootFolders(rootFolders);

        const qualityProfilesResponse = await fetch(`${import.meta.env.VITE_API_URL}/sonarr/profiles`);
        const qualityProfiles = await qualityProfilesResponse.json();
        setSonarrQualityProfiles(qualityProfiles);
      } catch (error) {
        console.error("Error fetching Sonarr data:", error);
      }
    };

    const fetchRadarrData = async () => {
      try {
        const rootFoldersResponse = await fetch(`${import.meta.env.VITE_API_URL}/radarr/rootfolders`);
        const rootFolders = await rootFoldersResponse.json();
        setRadarrRootFolders(rootFolders);

        const qualityProfilesResponse = await fetch(`${import.meta.env.VITE_API_URL}/radarr/profiles`);
        const qualityProfiles = await qualityProfilesResponse.json();
        setRadarrQualityProfiles(qualityProfiles);
      } catch (error) {
        console.error("Error fetching Radarr data:", error);
      }
    };

    if (selectedMode === "radarr-sonarr") {
      fetchSonarrData();
      fetchRadarrData();
    }
  }, [selectedMode]);


  const handleNext = () => {
    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1)
    } else {
      // Handle form submission here
      console.log("Form submitted", { traktClientId, traktClientSecret, selectedMode, settings })
    }
  }

  const handlePrevious = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1)
    }
  }

  const handleSettingChange = (key: string, value: string) => {
    setSettings((prevSettings) => ({ ...prevSettings, [key]: value }))
  }

  const renderStepContent = () => {
    switch (currentStep) {
      case 0:
        return (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="trakt-client-id">Trakt Client ID</Label>
              <Input
                id="trakt-client-id"
                value={traktClientId}
                onChange={(e) => setTraktClientId(e.target.value)}
                placeholder="Enter your Trakt Client ID"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="trakt-client-secret">Trakt Client Secret</Label>
              <Input
                id="trakt-client-secret"
                type="password"
                value={traktClientSecret}
                onChange={(e) => setTraktClientSecret(e.target.value)}
                placeholder="Enter your Trakt Client Secret"
              />
            </div>
          </div>
        )
      case 1:
        return (
          <RadioGroup value={selectedMode} onValueChange={setSelectedMode}>
            {modes.map((mode) => (
              <div key={mode.value} className="flex items-center space-x-2">
                <RadioGroupItem value={mode.value} id={mode.value} />
                <Label htmlFor={mode.value}>{mode.label}</Label>
              </div>
            ))}
          </RadioGroup>
        )
      case 2:
        return renderModeSettings()
      default:
        return null
    }
  }

  const renderModeSettings = () => {
    switch (selectedMode) {
      case "ombi":
        return (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="ombi-base-url">Ombi Base URL</Label>
              <Input
                id="ombi-base-url"
                value={settings["ombi-base-url"] || ""}
                onChange={(e) => handleSettingChange("ombi-base-url", e.target.value)}
                placeholder="Enter Ombi Base URL"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="ombi-api-key">Ombi API Key</Label>
              <Input
                id="ombi-api-key"
                type="password"
                value={settings["ombi-api-key"] || ""}
                onChange={(e) => handleSettingChange("ombi-api-key", e.target.value)}
                placeholder="Enter Ombi API Key"
              />
            </div>
          </div>
        )
      case "radarr-sonarr":
        return (
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="sonarr-base-url">Sonarr Base URL</Label>
                <Input
                  id="sonarr-base-url"
                  value={settings["sonarr-base-url"] || ""}
                  onChange={(e) => handleSettingChange("sonarr-base-url", e.target.value)}
                  placeholder="http://localhost:8989"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="sonarr-api-key">Sonarr API Key</Label>
                <Input
                  id="sonarr-api-key"
                  type="password"
                  value={settings["sonarr-api-key"] || ""}
                  onChange={(e) => handleSettingChange("sonarr-api-key", e.target.value)}
                  placeholder="Enter Sonarr API Key"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="sonarr-root-folder">Sonarr Root Folder</Label>
                <Select
                  onValueChange={(value) => handleSettingChange("sonarr-root-folder", value)}
                  value={settings["sonarr-root-folder"] || ""}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select a root folder to use..." />
                  </SelectTrigger>
                  <SelectContent>
                    {sonarrRootFolders.map((folder) => (
                      <SelectItem key={folder.id} value={folder.id.toString()}>
                        {folder.path} (Free space: {formatBytes(folder.free_space)})
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="sonarr-quality-profile">Sonarr Quality Profile</Label>
                <Select
                  onValueChange={(value) => handleSettingChange("sonarr-quality-profile", value)}
                  value={settings["sonarr-quality-profile"] || ""}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select a quality profile..." />
                  </SelectTrigger>
                  <SelectContent>
                    {sonarrQualityProfiles.map((profile) => (
                      <SelectItem key={profile.id} value={profile.id.toString()}>
                        {profile.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <Separator />

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="radarr-base-url">Radarr Base URL</Label>
                <Input
                  id="radarr-base-url"
                  value={settings["radarr-base-url"] || ""}
                  onChange={(e) => handleSettingChange("radarr-base-url", e.target.value)}
                  placeholder="http://localhost:7878"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="radarr-api-key">Radarr API Key</Label>
                <Input
                  id="radarr-api-key"
                  type="password"
                  value={settings["radarr-api-key"] || ""}
                  onChange={(e) => handleSettingChange("radarr-api-key", e.target.value)}
                  placeholder="Enter Radarr API Key"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="radarr-root-folder">Radarr Root Folder</Label>
                <Select
                  onValueChange={(value) => handleSettingChange("radarr-root-folder", value)}
                  value={settings["radarr-root-folder"] || ""}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select a root folder to use..." />
                  </SelectTrigger>
                  <SelectContent>
                    {radarrRootFolders.map((folder) => (
                      <SelectItem key={folder.id} value={folder.id.toString()}>
                        {folder.path} (Free space: {formatBytes(folder.free_space)})
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="radarr-quality-profile">Radarr Quality Profile</Label>
                <Select
                  onValueChange={(value) => handleSettingChange("radarr-quality-profile", value)}
                  value={settings["radarr-quality-profile"] || ""}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select a quality profile..." />
                  </SelectTrigger>
                  <SelectContent>
                    {radarrQualityProfiles.map((profile) => (
                      <SelectItem key={profile.id} value={profile.id.toString()}>
                        {profile.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>
        )
      default:
        return null
    }
  }

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
        <Button onClick={handleNext}>
          {currentStep === steps.length - 1 ? "Finish" : "Next"}
        </Button>
      </CardFooter>
    </Card>
  )
}