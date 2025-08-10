package model

// WithTemperature returns a copy of client with the given temperature if it is an OpenAI client.
func WithTemperature(c Client, t float64) Client {
	if oa, ok := c.(*OpenAI); ok {
		cp := *oa
		cp.Temperature = &t
		return &cp
	}
	return c
}
