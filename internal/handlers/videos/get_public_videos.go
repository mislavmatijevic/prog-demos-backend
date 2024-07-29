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
			Id:    1,
			Name:  "Kako Napisati C++ Program [1. dio - Osnove Jezika C++]",
			Link:  "https://youtu.be/RyL2MjxgVj0",
			Topic: models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:    2,
			Name:  "C++ Varijable [2. dio - Osnove Jezika C++]",
			Link:  "https://youtu.be/RcFVMaGdSKM",
			Topic: models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:    3,
			Name:  "C++ Logika [3. dio - Osnove Jezika C++]",
			Link:  "https://youtu.be/BqdPEeVPSB0",
			Topic: models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:    4,
			Name:  "C++ Petlje [4. dio - Osnove Jezika C++]",
			Link:  "https://youtu.be/3COiJ6b5sq4",
			Topic: models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:    5,
			Name:  "C++ Polja [1. demonstrature - PROG1]",
			Link:  "https://youtu.be/BSvFewITLv4",
			Topic: models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:    6,
			Name:  "C++ Sortiranja [2. demonstrature - PROG1]",
			Link:  "https://youtu.be/NYbzVk5vncM",
			Topic: models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:    7,
			Name:  "C++ Slogovi i Unije (1/3) [3. demonstrature - PROG1]",
			Link:  "https://youtu.be/fTXnvSbAbWE",
			Topic: models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:    8,
			Name:  "C++ Slogovi i Unije (2/3) [3. demonstrature - PROG1]",
			Link:  "https://youtu.be/WKoyPZxLOWM",
			Topic: models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:    9,
			Name:  "C++ Slogovi i Unije (3/3) [3. demonstrature - PROG1]",
			Link:  "https://youtu.be/-iLFn0Ttbbo",
			Topic: models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:    10,
			Name:  "C++ Pokazivači i Vezana Lista (1/3) [4. demonstrature - PROG1]",
			Link:  "https://youtu.be/oN-RparzioU",
			Topic: models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:    11,
			Name:  "C++ Pokazivači i Vezana Lista (2/3) [4. demonstrature - PROG1]",
			Link:  "https://youtu.be/JOWopUG_U4I",
			Topic: models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:    12,
			Name:  "C++ Pokazivači i Vezana Lista (3/3 dodatno o pokazivačima) [4. demonstrature - PROG1]",
			Link:  "https://youtu.be/F1sNQxkxfbg",
			Topic: models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:    13,
			Name:  "C++ Funkcije [5. demonstrature - PROG1]",
			Link:  "https://youtu.be/KDI61ExnKZs",
			Topic: models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:    15,
			Name:  "C++ Rekurzije [6. demonstrature - PROG1]",
			Link:  "https://youtu.be/8TI-NIByHR0",
			Topic: models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:    16,
			Name:  "C++ Tekstualne Datoteke [7. demonstrature - PROG1]",
			Link:  "https://youtu.be/4k_sr8_v75s",
			Topic: models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
		{
			Id:    17,
			Name:  "C++ Binarne Datoteke [8. demonstrature - PROG1]",
			Link:  "https://youtu.be/SmnwRqHMLuw",
			Topic: models.Topic{Id: 1, Name: "Osnove programiranja uz C++"}},
	},
}

func GetPublicVideos(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(videosResponse.Videos)
}
