package main

import (
	"fmt"
	"testing"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// La trampa: tratar un Output como si fuera un string.
func TestTrampaOutput(t *testing.T) {
	m := &mocks{tipos: map[string]int{}}
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		b, err := s3.NewBucket(ctx, "datos", &s3.BucketArgs{Bucket: pulumi.String("g680-datos-prod")})
		if err != nil {
			return err
		}
		mal := fmt.Sprintf("%v/*", b.Arn)
		t.Logf("fmt.Sprintf   → %q", mal)
		bien := pulumi.Sprintf("%s/*", b.Arn)
		bien.ApplyT(func(s string) string {
			t.Logf("pulumi.Sprintf → %q", s)
			return s
		})
		return nil
	}, pulumi.WithMocks("g680", "prod", m))
	if err != nil {
		t.Fatal(err)
	}
}
