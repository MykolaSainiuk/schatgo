package cmnerr

import "encoding/json"

var r, _ = json.Marshal(struct{}{})
var EmptyJSONResponse = r
