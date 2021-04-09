package pingdom

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nordcloud/go-pingdom/pingdom"
	"github.com/nordcloud/go-pingdom/solarwinds"
	"github.com/stretchr/testify/assert"
	"html/template"
	"strconv"
	"testing"
	"time"
)

func TestAccOccurrence_basic(t *testing.T) {
	resourceName := "pingdom_occurrence.test"
	occurrenceNum := 3
	resp := createNewMaintenance(t, time.Duration(occurrenceNum))
	group := OccurrenceGroup{
		MaintenanceId: int64(resp.ID),
		From:          resp.From,
		To:            resp.EffectiveTo,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOccurrenceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOccurrence_basicConfig(group, 0, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "from", strconv.FormatInt(resp.From, 10)),
					resource.TestCheckResourceAttr(resourceName, "to", strconv.FormatInt(resp.To, 10)),
					resource.TestCheckResourceAttr(resourceName, "size", strconv.Itoa(occurrenceNum+1)),
				),
			},
		},
	})
}

func TestAccOccurrence_update(t *testing.T) {
	resourceName := "pingdom_occurrence.test"
	occurrenceNum := 3
	resp := createNewMaintenance(t, time.Duration(occurrenceNum))
	group := OccurrenceGroup{
		MaintenanceId: int64(resp.ID),
		From:          resp.From,
		To:            resp.EffectiveTo,
	}

	from, to := resp.From, time.Unix(resp.To, 0).Add(1*time.Hour).Unix()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOccurrenceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOccurrence_basicConfig(group, 0, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "from", strconv.FormatInt(resp.From, 10)),
					resource.TestCheckResourceAttr(resourceName, "to", strconv.FormatInt(resp.To, 10)),
					resource.TestCheckResourceAttr(resourceName, "size", strconv.Itoa(occurrenceNum+1)),
				),
			},
			{
				Config: testAccOccurrence_basicConfig(group, from, to),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "from", strconv.FormatInt(from, 10)),
					resource.TestCheckResourceAttr(resourceName, "to", strconv.FormatInt(to, 10)),
					resource.TestCheckResourceAttr(resourceName, "size", strconv.Itoa(occurrenceNum+1)),
				),
			},
		},
	})
}

func testAccCheckOccurrenceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Clients).Pingdom

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "pingdom_occurrence" {
			continue
		}

		g := OccurrenceGroup{}
		if v, err := strconv.ParseInt(rs.Primary.Attributes["maintenance_id"], 10, 64); err != nil {
			return err
		} else {
			g.MaintenanceId = v
		}
		if v, err := strconv.ParseInt(rs.Primary.Attributes["effective_from"], 10, 64); err != nil {
			return err
		} else {
			g.From = v
		}
		if v, err := strconv.ParseInt(rs.Primary.Attributes["effective_to"], 10, 64); err != nil {
			return err
		} else {
			g.To = v
		}

		if size, err := g.Size(client); err != nil {
			return err
		} else if size != 0 {
			return fmt.Errorf("the occurrence has not been deleted, %d left", size)
		}

		_, err := client.Maintenances.Delete(int(g.MaintenanceId))
		if err != nil {
			return err
		}
	}
	return nil
}

func createNewMaintenance(t *testing.T, occurrenceNum time.Duration) *pingdom.MaintenanceResponse {
	now := time.Now()
	from := now.Add(1 * time.Hour)
	to := from.Add(1 * time.Hour)
	maintenance := pingdom.MaintenanceWindow{
		Description:    "terraform resource test - " + solarwinds.RandString(10),
		From:           from.Unix(),
		To:             to.Unix(),
		RecurrenceType: "day",
		RepeatEvery:    1,
		EffectiveTo:    to.Add(occurrenceNum * 24 * time.Hour).Unix(),
	}

	pingdomClient, err := pingdom.NewClientWithConfig(pingdom.ClientConfig{})
	assert.NoError(t, err)
	resp, err := pingdomClient.Maintenances.Create(&maintenance)
	assert.NoError(t, err)
	resp, err = pingdomClient.Maintenances.Read(resp.ID)
	assert.NoError(t, err)
	return resp
}

func testAccOccurrence_basicConfig(group OccurrenceGroup, from int64, to int64) string {
	t := template.Must(template.New("basicConfig").Parse(`
resource "pingdom_occurrence" "test" {
	{{with .group}}
	maintenance_id = {{.MaintenanceId}}
	effective_from = {{.From}}
	effective_to = {{.To}}
	{{end}}
	{{if .from}}
	from = {{.from}}
	{{end}}
	{{if .to}}
	to = {{.to}}
	{{end}}
}
`))
	var buf bytes.Buffer
	if err := t.Execute(&buf, map[string]interface{}{
		"group": group,
		"from":  from,
		"to":    to,
	}); err != nil {
		panic(err)
	}
	result := buf.String()
	return result
}
