package pingdom

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nordcloud/go-pingdom/pingdom"
	"log"
	"strconv"
	"strings"
)

func resourcePingdomOccurrences() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePingdomOccurrencesCreate,
		ReadContext:   resourcePingdomOccurrencesRead,
		UpdateContext: resourcePingdomOccurrencesUpdate,
		DeleteContext: resourcePingdomOccurrencesDelete,
		Schema: map[string]*schema.Schema{
			"maintenance_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"from": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: false,
			},
			"to": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: false,
			},
		},
	}
}

type OccurrenceGroup pingdom.ListOccurrenceQuery

func NewOccurrenceGroupWithResourceData(d *schema.ResourceData) (*OccurrenceGroup, error) {
	q := OccurrenceGroup{}

	// required
	if v, ok := d.GetOk("maintenance_id"); ok {
		q.MaintenanceId = int64(v.(int))
	}

	if v, ok := d.GetOk("from"); ok {
		q.From = int64(v.(int))
	}

	if v, ok := d.GetOk("to"); ok {
		q.To = int64(v.(int))
	}

	return &q, nil
}

func NewOccurrenceGroupWithId(id string) (*OccurrenceGroup, error) {
	tokens := strings.Split(id, "-")
	if len(tokens) != 3 {
		return nil, fmt.Errorf("invalid id %s, not enough tokens", id)
	}
	g := OccurrenceGroup{}
	if i, err := strconv.ParseInt(tokens[0], 10, 64); err != nil {
		return nil, err
	} else {
		g.MaintenanceId = i
	}
	if i, err := strconv.ParseInt(tokens[1], 10, 64); err != nil {
		return nil, err
	} else {
		g.From = i
	}
	if i, err := strconv.ParseInt(tokens[2], 10, 64); err != nil {
		return nil, err
	} else {
		g.To = i
	}
	return &g, nil
}

func (g *OccurrenceGroup) Id() string {
	return fmt.Sprintf("%d-%d-%d", g.MaintenanceId, g.From, g.To)
}

func (g *OccurrenceGroup) List(client *pingdom.Client) ([]pingdom.Occurrence, error) {
	return client.Occurrences.List(pingdom.ListOccurrenceQuery(*g))
}

func (g *OccurrenceGroup) Populate(client *pingdom.Client, d *schema.ResourceData) error {
	if sample, err := g.Sample(client); err != nil {
		return err
	} else {
		if err = d.Set("from", sample.From); err != nil {
			return err
		}
		if err = d.Set("to", sample.To); err != nil {
			return err
		}
	}
	return nil
}

func (g *OccurrenceGroup) Sample(client *pingdom.Client) (*pingdom.Occurrence, error) {
	occurrences, err := g.List(client)
	if err != nil {
		return nil, err
	} else if len(occurrences) == 0 {
		return nil, fmt.Errorf("there are no occurrences matching query: %#v", g)
	} else {
		return &occurrences[0], nil
	}
}

func (g *OccurrenceGroup) Size(client *pingdom.Client) (int, error) {
	occurrences, err := g.List(client)
	if err != nil {
		return 0, err
	}

	return len(occurrences), nil
}

func (g *OccurrenceGroup) MustExists(client *pingdom.Client) error {
	if size, err := g.Size(client); err != nil {
		return err
	} else if size == 0 {
		return fmt.Errorf("there are no occurrences matching query: %#v", g)
	} else {
		return nil
	}
}

func (g *OccurrenceGroup) Update(client *pingdom.Client, update OccurrenceGroup) error {
	occurrenceUpdate := pingdom.Occurrence{
		From: update.From,
		To:   update.To,
	}
	return g.groupOp(client, func(occurrence pingdom.Occurrence) (interface{}, error) {
		return client.Occurrences.Update(occurrence.Id, occurrenceUpdate)
	})
}

func (g *OccurrenceGroup) Delete(client *pingdom.Client) error {
	return g.groupOp(client, func(occurrence pingdom.Occurrence) (interface{}, error) {
		return client.Occurrences.Delete(occurrence.Id)
	})
}

func (g *OccurrenceGroup) groupOp(client *pingdom.Client, op func(occurrence pingdom.Occurrence) (interface{}, error)) error {
	occurrences, err := client.Occurrences.List(pingdom.ListOccurrenceQuery(*g))
	if err != nil {
		return err
	}

	cancelChan := make(chan bool)
	errChan := make(chan error, len(occurrences))
	for _, occurrence := range occurrences {
		go func(occurrence pingdom.Occurrence) {
			select {
			case <-cancelChan:
				return
			default:
				_, err := op(occurrence)
				errChan <- err
			}
		}(occurrence)
	}

	expectTotal := len(occurrences)
	count := 0
	for err := range errChan {
		if err != nil {
			close(cancelChan)
			return err
		} else {
			count += 1
		}
		if expectTotal == count {
			break
		}
	}
	return nil
}

func resourcePingdomOccurrencesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Clients).Pingdom

	g, err := NewOccurrenceGroupWithResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Retrieve occurrences with query: %#v", g)
	if err := g.Populate(client, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(g.Id())

	return nil
}

func resourcePingdomOccurrencesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Clients).Pingdom

	id := d.Id()
	g, err := NewOccurrenceGroupWithId(id)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := g.Populate(client, d); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourcePingdomOccurrencesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Clients).Pingdom

	id := d.Id()
	occurrence, err := NewOccurrenceGroupWithId(id)
	if err != nil {
		return diag.FromErr(err)
	}

	update, err := NewOccurrenceGroupWithResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Occurrence update configuration: %#v", update)

	if err := occurrence.Update(client, *update); err != nil {
		return diag.FromErr(err)
	}

	if err := occurrence.Populate(client, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourcePingdomOccurrencesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Clients).Pingdom

	occurrence, err := NewOccurrenceGroupWithId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = occurrence.Delete(client)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
