package provider

// ParseStatusJSON decodes a JSON status payload (recorded API response or hand-written fixture).
func ParseStatusJSON(data []byte) (NodeStatus, error) {
	var st NodeStatus
	if err := DecodeJSON(data, &st); err != nil {
		return NodeStatus{}, err
	}
	return st, nil
}
