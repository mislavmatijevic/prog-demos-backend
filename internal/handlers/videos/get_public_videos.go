package videos

import (
	"encoding/json"
	"net/http"

	models "github.com/mislavmatijevic/prog-demos-backend/internal/models"
)

var videosResponse = struct {
	Videos []models.Video
}{
	Videos: []models.Video{
		{
			Id:         1,
			Name:       "Kako Napisati C++ Program",
			Subtitle:   "1. dio - Osnove Jezika C++",
			Identifier: "RyL2MjxgVj0",
			Topic:      models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:         2,
			Name:       "C++ Varijable",
			Subtitle:   "2. dio - Osnove Jezika C++",
			Identifier: "RcFVMaGdSKM",
			Topic:      models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:         3,
			Name:       "C++ Logika",
			Subtitle:   "3. dio - Osnove Jezika C++",
			Identifier: "BqdPEeVPSB0",
			Topic:      models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:         4,
			Name:       "C++ Petlje",
			Subtitle:   "4. dio - Osnove Jezika C++",
			Identifier: "3COiJ6b5sq4",
			Topic:      models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:         5,
			Name:       "C++ Polja",
			Subtitle:   "1. demonstrature - PROG1",
			Identifier: "BSvFewITLv4",
			Topic:      models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:         6,
			Name:       "C++ Sortiranja",
			Subtitle:   "2. demonstrature - PROG1",
			Identifier: "NYbzVk5vncM",
			Topic:      models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:         7,
			Name:       "C++ Slogovi i Unije (1/3)",
			Subtitle:   "3. demonstrature - PROG1",
			Identifier: "fTXnvSbAbWE",
			Topic:      models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:         8,
			Name:       "C++ Slogovi i Unije (2/3)",
			Subtitle:   "3. demonstrature - PROG1",
			Identifier: "WKoyPZxLOWM",
			Topic:      models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:         9,
			Name:       "C++ Slogovi i Unije (3/3)",
			Subtitle:   "3. demonstrature - PROG1",
			Identifier: "-iLFn0Ttbbo",
			Topic:      models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:         10,
			Name:       "C++ Pokazivači i Vezana Lista (1/3)",
			Subtitle:   "4. demonstrature - PROG1",
			Identifier: "oN-RparzioU",
			Topic:      models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:         11,
			Name:       "C++ Pokazivači i Vezana Lista (2/3)",
			Subtitle:   "4. demonstrature - PROG1",
			Identifier: "JOWopUG_U4I",
			Topic:      models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:         12,
			Name:       "C++ Pokazivači i Vezana Lista (3/3 dodatno o pokazivačima)",
			Subtitle:   "4. demonstrature - PROG1",
			Identifier: "F1sNQxkxfbg",
			Topic:      models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:         13,
			Name:       "C++ Funkcije",
			Subtitle:   "5. demonstrature - PROG1",
			Identifier: "KDI61ExnKZs",
			Topic:      models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:         15,
			Name:       "C++ Rekurzije",
			Subtitle:   "6. demonstrature - PROG1",
			Identifier: "8TI-NIByHR0",
			Topic:      models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:         16,
			Name:       "C++ Tekstualne Datoteke",
			Subtitle:   "7. demonstrature - PROG1",
			Identifier: "4k_sr8_v75s",
			Topic:      models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:         17,
			Name:       "C++ Binarne Datoteke",
			Subtitle:   "8. demonstrature - PROG1",
			Identifier: "SmnwRqHMLuw",
			Topic:      models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
	},
}

func GetPublicVideos(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(videosResponse.Videos)
}
