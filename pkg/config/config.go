package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	ApiTokenField = field.StringField(
		"api-token",
		field.WithDescription("The Percipio Bearer Token"),
		field.WithRequired(true),
	)
	OrganizationIdField = field.StringField(
		"organization-id",
		field.WithDescription("The Percipio Organization ID"),
		field.WithRequired(true),
	)
	LookbackDaysField = field.IntField(
		"lookback-days",
		field.WithDescription("How many days back of learning activity data to fetch"),
		field.WithShortHand("d"),
	)
	LookbackYearsField = field.IntField(
		"lookback-years",
		field.WithDescription("How many years back of learning activity data to fetch (default: 10)"),
		field.WithShortHand("y"),
		field.WithDefaultValue(10),
	)

	// ConfigurationFields defines the external configuration required for the
	// connector to run. Note: these fields can be marked as optional or
	// required.
	ConfigurationFields = []field.SchemaField{
		ApiTokenField,
		OrganizationIdField,
		LookbackDaysField,
		LookbackYearsField,
	}

	// FieldRelationships defines relationships between the fields listed in
	// ConfigurationFields that can be automatically validated. For example, a
	// username and password can be required together, or an access token can be
	// marked as mutually exclusive from the username password pair.
	FieldRelationships []field.SchemaFieldRelationship = nil

	ConfigurationSchema = field.NewConfiguration(
		ConfigurationFields,
	)
)
