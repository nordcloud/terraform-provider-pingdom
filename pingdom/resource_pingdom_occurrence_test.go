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
		To:            resp.To,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOccurrenceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOccurrence_basicConfig(group),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "from", strconv.FormatInt(group.From, 10)),
					resource.TestCheckResourceAttr(resourceName, "to", strconv.FormatInt(group.To, 10)),
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
		To:            resp.To,
	}

	update := group
	update.To = time.Unix(update.To, 0).Add(1 * time.Hour).Unix()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOccurrenceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOccurrence_basicConfig(group),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "from", strconv.FormatInt(group.From, 10)),
					resource.TestCheckResourceAttr(resourceName, "to", strconv.FormatInt(group.To, 10)),
				),
			},
			{
				Config: testAccOccurrence_basicConfig(update),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "from", strconv.FormatInt(update.From, 10)),
					resource.TestCheckResourceAttr(resourceName, "to", strconv.FormatInt(update.To, 10)),
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

		id := rs.Primary.ID
		g, err := NewOccurrenceGroupWithId(id)
		if err != nil {
			return err
		}

		if size, err := g.Size(client); err != nil {
			return err
		} else if size != 0 {
			return fmt.Errorf("the occurrence has not been deleted, %d left", size)
		}

		_, err = client.Maintenances.Delete(int(g.MaintenanceId))
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

func testAccOccurrence_basicConfig(group OccurrenceGroup) string {
	t := template.Must(template.New("basicConfig").Parse(`
resource "pingdom_occurrence" "test" {
	maintenance_id = {{.MaintenanceId}}
	from = {{.From}}
	to = {{.To}}
}
`))
	var buf bytes.Buffer
	if err := t.Execute(&buf, group); err != nil {
		panic(err)
	}
	result := buf.String()
	return result
}
