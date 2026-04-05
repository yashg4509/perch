package config

// Load parses and validates perch.yaml bytes.
func Load(data []byte) (*Config, error) {
	c, err := Parse(data)
	if err != nil {
		return nil, err
	}
	if err := Validate(c); err != nil {
		return nil, err
	}
	return c, nil
}
