package getstream

import "encoding/json"

func prepareForGetstream(activity *FlatFeedActivity) ([]byte, error) {

	payload := make(map[string]json.RawMessage)

	for key, value := range activity.MetaData {
		payload[key] = raw(value)
	}

	payload["actor"] = raw(activity.Actor)
	payload["verb"] = raw(activity.Verb)
	payload["object"] = raw(activity.Object)

	if activity.ID != "" {
		payload["id"] = raw(activity.ID)
	}
	if activity.Target != "" {
		payload["target"] = raw(activity.Target)
	}
	if activity.TimeStamp != nil {
		payload["time"] = raw(activity.TimeStamp)
	}
	if activity.ForeignID != "" {
		payload["foreign_id"] = raw(activity.ForeignID)
	}
	if activity.Data != nil {
		payload["data"] = raw(activity.Data)
	}
	if len(activity.To) > 0 {
		payload["to"] = raw(activity.To)
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func raw(input interface{}) json.RawMessage {

	if input == nil {
		return nil
	}

	b, err := json.Marshal(input)
	if err != nil {
		return nil
	}
	return json.RawMessage(b)
}

func extractFromGetStream(payload []byte) *postFlatFeedOutputActivity {

	activity := postFlatFeedOutputActivity{}
	rawPayload := make(map[string]json.RawMessage)
	metadata := make(map[string]string)

	json.Unmarshal(payload, &activity)
	json.Unmarshal(payload, &rawPayload)

	for key, value := range rawPayload {

		if key != "id" && key != "actor" && key != "verb" && key != "object" && key != "target" && key != "time" && key != "foreign_id" && key != "data" && key != "to" {
			var strValue string
			json.Unmarshal(value, strValue)
			metadata[key] = strValue
		}
	}

	activity.MetaData = metadata
	return &activity
}
