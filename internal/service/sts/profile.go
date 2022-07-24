package sts

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	profileNameRe     = regexp.MustCompile(`\[|]`)
	awsAccessRe       = regexp.MustCompile(`aws_access_key_id\s*=\s*`)
	awsSecretAccessRe = regexp.MustCompile(`aws_secret_access_key\s*=\s*`)
	awsSessionRe      = regexp.MustCompile(`aws_session_token\s*=\s*`)
)

// profile, temporary struct to handle the parsing of the configured profiles
type profile struct {
	Name               string
	AWSAccessKeyId     string
	AWSSecretAccessKey string `type:"string" sensitive:"true"`
	AWSSessionToken    string `type:"string" sensitive:"true"`
	Region             string
}

func DataSourceProfile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceProfilesRead,

		Schema: map[string]*schema.Schema{
			"profiles": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"aws_access_key_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"aws_secret_access_key": {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
						"aws_session_token": {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
						"region": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceProfilesRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Reading profiles from home credentials file")
	profiles, err := getProfiles()

	if err != nil {
		return fmt.Errorf("getting Profiles: %w", err)
	}

	if len(profiles) == 0 {
		return fmt.Errorf("No profiles found. Setup profiles before calling this.")
	}

	profilesFlattened := flattenProfiles(profiles)

	if err := d.Set("profiles", profilesFlattened); err != nil {
		return fmt.Errorf("Error setting result: %w", err)
	}

	d.SetId(createIdFromProfileNames(profiles))

	return nil
}

func flattenProfiles(profiles []profile) (profs []interface{}) {
	profs = make([]interface{}, len(profiles), len(profiles))

	for i, profile := range profiles {
		pr := make(map[string]interface{})
		pr["name"] = profile.Name
		pr["aws_access_key_id"] = profile.AWSAccessKeyId
		pr["aws_secret_access_key"] = profile.AWSSecretAccessKey
		pr["aws_session_token"] = profile.AWSSessionToken
		pr["region"] = profile.Region

		profs[i] = pr
	}
	return profs
}

func createIdFromProfileNames(profiles []profile) (id string) {
	for _, profile := range profiles {
		id = id + "_" + profile.Name
	}
	return id
}

func getProfiles() ([]profile, error) {
	profiles := []profile{}
	credentialsFilePath := getAWSCredentialsFilePath()
	file, err := os.Open(credentialsFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// optionally, resize scanner's capacity for lines over 64K, see next example
	var currentProfile *profile
	for scanner.Scan() {
		currentLine := scanner.Text()
		if strings.Contains(currentLine, "[") {
			if currentProfile == nil {
				currentProfile = new(profile)
			} else {
				profiles = append(profiles, *currentProfile)
				currentProfile = new(profile)
			}
			currentProfile.Name = profileNameRe.ReplaceAllString(currentLine, "")
		} else if strings.Contains(currentLine, "aws_access_key_id") {
			currentProfile.AWSAccessKeyId = awsAccessRe.ReplaceAllString(currentLine, "")
		} else if strings.Contains(currentLine, "aws_secret_access_key") {
			currentProfile.AWSSecretAccessKey = awsSecretAccessRe.ReplaceAllString(currentLine, "")
		} else if strings.Contains(currentLine, "aws_session_token") {
			currentProfile.AWSSessionToken = awsSessionRe.ReplaceAllString(currentLine, "")
		} else if currentLine == "" {
			profiles = append(profiles, *currentProfile)
			currentProfile = nil
		}
	}
	if currentProfile != nil {
		profiles = append(profiles, *currentProfile)
		currentProfile = nil
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return profiles, nil
}

func getAWSCredentialsFilePath() string {
	homeDir, _ := os.UserHomeDir()
	return path.Join(homeDir, ".aws", "credentials")
}
