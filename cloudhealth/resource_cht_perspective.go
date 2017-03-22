package cloudhealth

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

const apiUrl string = "https://chapi.cloudhealthtech.com/v1/perspective_schemas"

func resourceCHTPerspective() *schema.Resource {
	return &schema.Resource{
		Create: resourceCHTPerspectiveCreate,
		Read:   resourceCHTPerspectiveRead,
		Update: resourceCHTPerspectiveUpdate,
		Delete: resourceCHTPerspectiveDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"include_in_reports": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: false,
			},
			"group": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ForceNew: false,
						},
						"ref_id": &schema.Schema{
							Type:     schema.TypeString,
							ForceNew: false,
							Computed: true,
							Optional: true,
						},
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: false,
							Default:  "filter",
						},
						"rule": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: false,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"asset": &schema.Schema{
										Type:     schema.TypeString,
										Required: true,
										ForceNew: false,
									},
									// for type="categorize"
									"tag_field": &schema.Schema{
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: false,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									// for type="categorize"
									"field": &schema.Schema{
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: false,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"combine_with": &schema.Schema{
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: false,
									},
									"condition": &schema.Schema{
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: false,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"tag_field": &schema.Schema{
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: false,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"field": &schema.Schema{
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: false,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"op": &schema.Schema{
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: false,
													Default:  "=",
												},
												"val": &schema.Schema{
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: false,
												},
											},
										},
									},
								},
							},
						},
						"dynamic_group": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: false,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ref_id": &schema.Schema{
										Type:     schema.TypeString,
										ForceNew: false,
										Computed: true,
										Optional: true,
									},
									"name": &schema.Schema{
										Type:     schema.TypeString,
										ForceNew: false,
										Computed: true,
										Optional: true,
									},
									"val": &schema.Schema{
										Type:     schema.TypeString,
										ForceNew: false,
										Computed: true,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"other_group": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"constant_type": &schema.Schema{
							Type:     schema.TypeString,
							ForceNew: false,
							Computed: true,
						},
						"ref_id": &schema.Schema{
							Type:     schema.TypeString,
							ForceNew: false,
							Computed: true,
						},
						"blk_id": &schema.Schema{
							Type:     schema.TypeString,
							ForceNew: false,
							Computed: true,
							Optional: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							ForceNew: false,
							Computed: true,
							Optional: true,
						},
						"val": &schema.Schema{
							Type:     schema.TypeString,
							ForceNew: false,
							Computed: true,
							Optional: true,
						},
						"is_other": &schema.Schema{
							Type:     schema.TypeString,
							ForceNew: false,
							Computed: true,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceCHTPerspectiveCreate(d *schema.ResourceData, meta interface{}) error {
	key := meta.(*ChtMeta).apiKey

	pj, err := tfToJson(d)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s?api_key=%s", apiUrl, key)
	resp, err := http.Post(url, "application/json", bytes.NewReader(pj))
	if err != nil {
		return fmt.Errorf("Failed to create perspective because %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to create perspective %s because got status code %d", d.Id(), resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to create perspective because %s", err)
	}
	re := regexp.MustCompile(`Perspective (\d*) created`)
	match := re.FindStringSubmatch(string(body))
	if match == nil || len(match) != 2 {
		return fmt.Errorf("Created perspective but didn't understand response to extract ID: %s", body)
	}
	d.SetId(match[1])

	return nil
}

func resourceCHTPerspectiveRead(d *schema.ResourceData, meta interface{}) error {
	key := meta.(*ChtMeta).apiKey

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Failed to parse %s as int because %s", d.Id(), err)
	}

	url := fmt.Sprintf("%s/%d?api_key=%s", apiUrl, id, key)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Failed to load perspective %s because %s", d.Id(), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to load perspective %s because got status code %d", d.Id(), resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read perspective %s because %s", d.Id(), err)
	}

	return jsonToTF(body, d)
}

func resourceCHTPerspectiveUpdate(d *schema.ResourceData, meta interface{}) error {
	key := meta.(*ChtMeta).apiKey
	pj, err := tfToJson(d)
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Failed to parse %s as int because %s", d.Id(), err)
	}

	url := fmt.Sprintf("%s/%d?api_key=%s", apiUrl, id, key)

	req, err := http.NewRequest("PUT", url, bytes.NewReader(pj))
	if err != nil {
		return fmt.Errorf("Failed to update perspective %s because %s", d.Id(), err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to update perspective %s because %s", d.Id(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to update perspective %s because got status code %d", d.Id(), resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)
	log.Println("Response to Cloudhealth PUT is:", bodyStr)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Got status code %d when attempting to update perspective: %s", resp.StatusCode, bodyStr)
	}

	return nil
}

func resourceCHTPerspectiveDelete(d *schema.ResourceData, meta interface{}) error {
	key := meta.(*ChtMeta).apiKey

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Failed to parse %s as int because %s", d.Id(), err)
	}

	url := fmt.Sprintf("%s/%d?api_key=%s", apiUrl, id, key)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("Failed to delete perspective %s because %s", d.Id(), err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to delete perspective %s because %s", d.Id(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to delete perspective %s because got status code %d", d.Id(), resp.StatusCode)
	}

	return nil
}
