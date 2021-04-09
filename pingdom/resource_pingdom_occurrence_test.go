package pingdom

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nordcloud/go-pingdom/pingdom"
	"github.com/nordcloud/go-pingdom/solarwinds"
	"html/template"
	"strconv"
	"testing"
	"time"
)

func TestAccOccurrence_basic(t *testing.T) {
	occurrenceNum := 3
	maintenance := getMaintenance(time.Duration(occurrenceNum))
	resourceName := "pingdom_occurrence.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOccurrenceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOccurrence_basicConfig(*maintenance, 0, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "from", strconv.FormatInt(maintenance.From, 10)),
					resource.TestCheckResourceAttr(resourceName, "to", strconv.FormatInt(maintenance.To, 10)),
					resource.TestCheckResourceAttr(resourceName, "size", strconv.Itoa(occurrenceNum+1)),
				),
			},
		},
	})
}

func TestAccOccurrence_update(t *testing.T) {
	occurrenceNum := 3
	maintenance := getMaintenance(time.Duration(occurrenceNum))
	resourceName := "pingdom_occurrence.test"

	from, to := maintenance.From, time.Unix(maintenance.To, 0).Add(1*time.Hour).Unix()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOccurrenceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOccurrence_basicConfig(*maintenance, 0, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "from", strconv.FormatInt(maintenance.From, 10)),
					resource.TestCheckResourceAttr(resourceName, "to", strconv.FormatInt(maintenance.To, 10)),
					resource.TestCheckResourceAttr(resourceName, "size", strconv.Itoa(occurrenceNum+1)),
				),
			},
			{
				Config: testAccOccurrence_basicConfig(*maintenance, from, to),
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
	}
	return nil
}

func getMaintenance(occurrenceNum time.Duration) *pingdom.MaintenanceWindow {
	now := time.Now()
	from := now.Add(1 * time.Hour)
	to := from.Add(1 * time.Hour)
	return &pingdom.MaintenanceWindow{
		Description:    "terraform resource test - " + solarwinds.RandString(10),
		From:           from.Unix(),
		To:             to.Unix(),
		RecurrenceType: "day",
		RepeatEvery:    1,
		EffectiveTo:    to.Add(occurrenceNum * 24 * time.Hour).Unix(),
	}
}

func testAccOccurrence_basicConfig(maintenance pingdom.MaintenanceWindow, from int64, to int64) string {
	t := template.Must(template.New("basicConfig").Parse(`
{{with .maintenance}}
resource "pingdom_maintenance" "test" {
	description = "{{.Description}}"
	from = {{.From}}
	to = {{.To}}
	recurrencetype = "day"
	repeatevery = 1
	effectiveto = {{.EffectiveTo}}
}
{{end}}

resource "pingdom_occurrence" "test" {
	maintenance_id = pingdom_maintenance.test.id
	effective_from = pingdom_maintenance.test.from
	effective_to = pingdom_maintenance.test.effectiveto
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
		"maintenance": maintenance,
		"from":        from,
		"to":          to,
	}); err != nil {
		panic(err)
	}
	result := buf.String()
	return result
}
