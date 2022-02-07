package producer

import (
	"context"
	"fmt"
	"github.com/sfomuseum/go-timings"
	"net/url"
)

type MultiProducer struct {
	Producer
	producers []Producer
}

func init() {
	ctx := context.Background()
	RegisterProducer(ctx, "multi", NewMultiProducer)
}

func NewMultiProducer(ctx context.Context, uri string) (Producer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	producer_uris := q["producer"]
	producers_count := len(producer_uris)

	if producers_count == 0 {
		return nil, fmt.Errorf("No ?producer parameters defined")
	}

	producers := make([]Producer, producers_count)

	for idx, u := range producer_uris {

		p, err := NewProducer(ctx, u)

		if err != nil {
			return nil, fmt.Errorf("Failed to create new producer for '%s', %w", u, err)
		}

		producers[idx] = p
	}

	p := &MultiProducer{
		producers: producers,
	}

	return p, nil
}

func (p *MultiProducer) PopulateWithIterator(ctx context.Context, monitor timings.Monitor, iterator_uri string, iterator_sources ...string) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err_ch := make(chan error)
	done_ch := make(chan bool)

	remaining := len(p.producers)

	for _, child_p := range p.producers {

		go func(child_p Producer) {

			err := child_p.PopulateWithIterator(ctx, monitor, iterator_uri, iterator_sources...)

			if err != nil {
				err_ch <- fmt.Errorf("Failed to iterator with %T, %w", child_p, err)
			}

			done_ch <- true

		}(child_p)

	}

	for remaining > 0 {

		select {
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			return err
		}
	}

	return nil
}

func (p *MultiProducer) Close(ctx context.Context) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err_ch := make(chan error)
	done_ch := make(chan bool)

	remaining := len(p.producers)

	for _, child_p := range p.producers {

		go func(child_p Producer) {

			err := child_p.Close(ctx)

			if err != nil {
				err_ch <- fmt.Errorf("Failed to close with %T, %w", child_p, err)
			}

			done_ch <- true

		}(child_p)

	}

	for remaining > 0 {

		select {
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			return err
		}
	}

	return nil
}
