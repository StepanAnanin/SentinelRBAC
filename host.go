package rbac

import (
	"errors"
)

// Host originaly desined for applications with microservice architectures.
//
// Host helps to define roles and schemas for each service in your app.
// You can also select several roles as default roles, all new users must have this roles.
type Host struct {
	// (Optional)
    //
    // Roles wich will have all new users, each default role must correspond with one of existing global roles.
    DefaultRoles []Role
    Roles        []Role
    Schemas      []Schema
}

func (h *Host) GetSchema(ID string) (*Schema, *Error) {
    if h == nil {
        return nil, NewError("RBAC schema is not defined")
    }

    if ID == "" {
        return nil, NewError("Missing schema id")
    }

    for _, schema := range h.Schemas {
        if schema.ID == ID {
            return &schema, nil
        }
    }

    return nil, NewError("Schema with id \"" + ID + "\" wasn't found")
}

func (h *Host) Validate() error {
	debugLog("[ RBAC ] Validating host...")

	if len(h.Schemas) == 0 {
		return errors.New("At least one schema must be defined")
	}

    if err := validateDefaultRoles(h.Roles, h.DefaultRoles); err != nil {
        return err
    }

    debugLog("[ RBAC ] Validating host: OK")

    return nil
}

// Merges permissions from schema specific roles with global roles.
// If any schema have a role with the same name as one of the global roles, then for each that role
// permissions of the schemas specific roles will overwrite permissions of the global roles.
// Also adds in schemas all global roles that wasn't explicitly specified for them.
func (h *Host) MergeRoles() {
	debugLog("[ RBAC ] Merging Host permissions of global and schemas roles...")

    schemas := make([]Schema, len(h.Schemas))

	for i, oldSchema := range h.Schemas {
        schema := oldSchema

        if oldSchema.Roles == nil || len(oldSchema.Roles) == 0 {
            schema.Roles = h.Roles
            schemas[i] = schema
            continue
        }

        roles := []Role{}

		for _, schemaRole := range schema.Roles {
			for _, role := range h.Roles {
				if schemaRole.Name == role.Name {
					roles = append(roles, schemaRole)
				} else {
					roles = append(roles, role)
				}
			}
		}

		schema.Roles = roles

        schemas[i] = schema
	}

    h.Schemas = schemas

    debugLog("[ RBAC ] Merging Host permissions of global and schemas roles: OK")
}

// Reads RBAC host configuration file from the given path.
// After loading and normalizing, validates this Host and returns an error if any of them were detected.
// Also merges permissions of the schema specific roles with permissions of the global roles.
func LoadHost(path string) (Host, error) {
    var zero Host

    raw, err := load[rawHost](path)
    if err != nil {
        return zero, err
    }

    host := raw.Normalize()

	if err = host.Validate(); err != nil {
		return zero, err
	}

	host.MergeRoles()

	return *host, nil
}

