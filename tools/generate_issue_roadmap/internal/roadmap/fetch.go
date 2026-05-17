package roadmap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func fetchProjectByGH(cfg config) (*projectData, error) {
	userProj, userErr := fetchProjectByGHPaged(cfg, true)
	if userErr == nil && userProj != nil {
		return userProj, nil
	}

	orgProj, orgErr := fetchProjectByGHPaged(cfg, false)
	if orgErr == nil && orgProj != nil {
		return orgProj, nil
	}

	if userErr != nil && orgErr != nil {
		return nil, fmt.Errorf("gh api 実行失敗 (user=%v, org=%v)", userErr, orgErr)
	}
	return nil, errors.New("projectV2 が見つかりませんでした")
}

func fetchProjectByGHPaged(cfg config, isUser bool) (*projectData, error) {
	query := graphQLQueryOrg
	if isUser {
		query = graphQLQueryUser
	}

	after := ""
	seenCursor := map[string]struct{}{}
	allItems := make([]projectItemNode, 0, 128)
	projectTitle := ""

	for {
		payload, err := runGraphQLQuery(cfg.Owner, cfg.ProjectNumber, query, after)
		if err != nil {
			return nil, err
		}

		page := parseProjectPageFromPayload(payload, isUser)
		if page == nil {
			return nil, nil
		}
		if projectTitle == "" {
			projectTitle = page.Title
		}
		allItems = append(allItems, page.Items...)

		if !page.HasNext || strings.TrimSpace(page.EndCursor) == "" {
			break
		}
		if _, exists := seenCursor[page.EndCursor]; exists {
			break
		}
		seenCursor[page.EndCursor] = struct{}{}
		after = page.EndCursor
	}

	return &projectData{Title: projectTitle, Items: allItems}, nil
}

func runGraphQLQuery(owner string, projectNumber int, query string, after string) ([]byte, error) {
	cmd := exec.Command(
		"gh", "api", "graphql",
		"-F", "owner="+owner,
		"-F", fmt.Sprintf("number=%d", projectNumber),
		"-f", "query="+query,
	)
	if strings.TrimSpace(after) != "" {
		cmd.Args = append(cmd.Args, "-F", "after="+after)
	}
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return nil, err
	}
	return out.Bytes(), nil
}

func parseProjectPageFromPayload(payload []byte, user bool) *projectPage {
	var resp struct {
		Data struct {
			User struct {
				ProjectV2 *struct {
					Title string `json:"title"`
					Items struct {
						Nodes    []projectItemNode `json:"nodes"`
						PageInfo struct {
							HasNextPage bool   `json:"hasNextPage"`
							EndCursor   string `json:"endCursor"`
						} `json:"pageInfo"`
					} `json:"items"`
				} `json:"projectV2"`
			} `json:"user"`
			Organization struct {
				ProjectV2 *struct {
					Title string `json:"title"`
					Items struct {
						Nodes    []projectItemNode `json:"nodes"`
						PageInfo struct {
							HasNextPage bool   `json:"hasNextPage"`
							EndCursor   string `json:"endCursor"`
						} `json:"pageInfo"`
					} `json:"items"`
				} `json:"projectV2"`
			} `json:"organization"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &resp); err != nil {
		return nil
	}
	if user {
		if resp.Data.User.ProjectV2 == nil {
			return nil
		}
		return &projectPage{
			Title:     resp.Data.User.ProjectV2.Title,
			Items:     resp.Data.User.ProjectV2.Items.Nodes,
			HasNext:   resp.Data.User.ProjectV2.Items.PageInfo.HasNextPage,
			EndCursor: resp.Data.User.ProjectV2.Items.PageInfo.EndCursor,
		}
	}
	if resp.Data.Organization.ProjectV2 == nil {
		return nil
	}
	return &projectPage{
		Title:     resp.Data.Organization.ProjectV2.Title,
		Items:     resp.Data.Organization.ProjectV2.Items.Nodes,
		HasNext:   resp.Data.Organization.ProjectV2.Items.PageInfo.HasNextPage,
		EndCursor: resp.Data.Organization.ProjectV2.Items.PageInfo.EndCursor,
	}
}

const graphQLQueryUser = `query($owner: String!, $number: Int!, $after: String) {
  user(login: $owner) {
    projectV2(number: $number) {
      title
      items(first: 100, after: $after) {
        nodes {
          content {
            ... on Issue {
              number
              title
              url
              repository {
                nameWithOwner
              }
            }
          }
          fieldValues(first: 30) {
            nodes {
              ... on ProjectV2ItemFieldSingleSelectValue {
                name
                field {
                  ... on ProjectV2SingleSelectField {
                    name
                  }
                }
              }
              ... on ProjectV2ItemFieldTextValue {
                text
                field {
                  ... on ProjectV2FieldCommon {
                    name
                  }
                }
              }
              ... on ProjectV2ItemFieldIterationValue {
                title
                field {
                  ... on ProjectV2IterationField {
                    name
                  }
                }
              }
            }
          }
        }
        pageInfo {
          hasNextPage
          endCursor
        }
      }
    }
  }
}`

const graphQLQueryOrg = `query($owner: String!, $number: Int!, $after: String) {
  organization(login: $owner) {
    projectV2(number: $number) {
      title
      items(first: 100, after: $after) {
        nodes {
          content {
            ... on Issue {
              number
              title
              url
              repository {
                nameWithOwner
              }
            }
          }
          fieldValues(first: 30) {
            nodes {
              ... on ProjectV2ItemFieldSingleSelectValue {
                name
                field {
                  ... on ProjectV2SingleSelectField {
                    name
                  }
                }
              }
              ... on ProjectV2ItemFieldTextValue {
                text
                field {
                  ... on ProjectV2FieldCommon {
                    name
                  }
                }
              }
              ... on ProjectV2ItemFieldIterationValue {
                title
                field {
                  ... on ProjectV2IterationField {
                    name
                  }
                }
              }
            }
          }
        }
        pageInfo {
          hasNextPage
          endCursor
        }
      }
    }
  }
}`
