package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	cfg "github.com/iiiatthew/baton-percipio-report/pkg/config"
	"github.com/iiiatthew/baton-percipio-report/pkg/connector"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	connectorName = "baton-percipio-report"
	version       = "dev"
)

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		connectorName,
		getConnector,
		cfg.ConfigurationSchema,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func getConnector(ctx context.Context, v *viper.Viper) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	// Parse the lookback duration with priority: days > years
	var lookbackDuration time.Duration
	
	lookbackDays := v.GetInt(cfg.LookbackDaysField.FieldName)
	lookbackYears := v.GetInt(cfg.LookbackYearsField.FieldName)
	
	switch {
	case lookbackDays > 0:
		// Convert days to hours (24 hours per day)
		lookbackDuration = time.Duration(lookbackDays) * 24 * time.Hour
		l.Info("Using days-based lookback", zap.Int("days", lookbackDays), zap.Duration("duration", lookbackDuration))
	case lookbackYears > 0:
		// Convert years to hours (365 days * 24 hours per year)
		lookbackDuration = time.Duration(lookbackYears) * 365 * 24 * time.Hour
		l.Info("Using years-based lookback", zap.Int("years", lookbackYears), zap.Duration("duration", lookbackDuration))
	default:
		// This shouldn't happen since years has a default value of 10
		lookbackDuration = 10 * 365 * 24 * time.Hour
		l.Info("Using default lookback (10 years)", zap.Duration("duration", lookbackDuration))
	}

	cb, err := connector.New(
		ctx,
		v.GetString(cfg.OrganizationIdField.FieldName),
		v.GetString(cfg.ApiTokenField.FieldName),
		lookbackDuration,
	)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}
	connector, err := connectorbuilder.NewConnector(ctx, cb)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}
	return connector, nil
}
