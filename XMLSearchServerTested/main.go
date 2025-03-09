package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
)

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if token := r.Header.Get("AccessToken"); token != "Отвечааааай!" {
		http.Error(w, fmt.Sprintf("Неверный токен: %s", token), http.StatusUnauthorized)
		return
	}
	const dataDir = "dataset.xml"
	type User struct {
		Id         int    `xml:"id"`
		First_name string `xml:"first_name" json:"-"`
		Last_name  string `xml:"last_name" json:"-"`
		Name       string `xml:"-"`
		Age        int    `xml:"age"`
		About      string `xml:"about"`
		Gender     string `xml:"gender"`
	}
	type xmlData struct {
		XMLName xml.Name `xml:"root"`
		Items   []User   `xml:"row"`
	}
	type possibleOrderFields struct {
		id   string
		age  string
		name string
	}
	fields := possibleOrderFields{id: `Id`, age: `Age`, name: `Name`}
	//http://127.0.0.1:53784/?query=cillum&order_field=Id&order_by=-1&limit=10&offset=0

	// что искать. Ищем по полям записи `Name` и `About` просто подстроку, без регулярок.
	// `Name` - это first_name + last_name из xml (вам надо руками пройтись в цикле по записям и сделать такой,
	// автоматом нельзя). Если поле пустое - то возвращаем все записи (поиск пустой подстроки всегда возвращает true),
	// т.е. делаем только логику сортировки
	query := r.FormValue("query")
	// по какому полю сортировать. Работает по полям `Id`, `Age`, `Name`, если пустой - то сортируем по `Name`, если что-то
	// другое - SearchServer ругается ошибкой.
	order_field := r.FormValue("order_field")
	switch order_field {
	case "":
		order_field = fields.name
	case fields.id, fields.age, fields.name:

	default:
		http.Error(w, fmt.Sprintf("Нет поля %s для сортировки", order_field), http.StatusBadRequest)
		return
	}
	// направление сортировки (как есть, по убыванию, по возрастанию), в client.go есть соответствующие константы
	str_order_by := r.FormValue("order_by")
	order_by, err := strconv.Atoi(str_order_by)
	if err != nil {
		http.Error(w, fmt.Sprintf("Порядок сортировки: ожидалось -1, 0 или 1, получено: %s", str_order_by), http.StatusBadRequest)
		return
	} else if order_by < -1 || order_by > 1 {
		http.Error(w, fmt.Sprintf("Порядок сортировки: ожидалось -1, 0 или 1, получено: %d", order_by), http.StatusBadRequest)
		return
	}
	// сколько записей вернуть
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil || limit < 0 {
		http.Error(w, fmt.Sprintf("Лимит: ожидалось целое положительное число, получено: %v", limit), http.StatusBadRequest)
		return
	}
	// начиня с какой записи вернуть (сколько пропустить с начала) - нужно для огранизации постраничной навигации
	offset, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil || offset < 0 {
		http.Error(w, fmt.Sprintf("Сколько пропустить: ожидалось целое положительное число, получено: %v", limit), http.StatusBadRequest)
		return
	}

	file, err := os.ReadFile(dataDir)
	if err != nil {
		fmt.Println("Ошибка открытия ", err)
		http.Error(w, fmt.Sprintln("Ошибка открытия", err), http.StatusInternalServerError)
		return
	}

	//data := make([]User, 0, currentNumberOfRecords)
	var data xmlData
	err = xml.Unmarshal(file, &data)
	if err != nil {
		fmt.Println("Ошибка XML", err)
		http.Error(w, fmt.Sprintln("Ошибка XML", err), http.StatusInternalServerError)
		return
	}

	var dataMap map[string]User
	switch query {
	case fields.id, fields.age:
		dataMap = make(map[string]User)
		//dataMap = dataMap.(map[int]User)
	default:
		dataMap = make(map[string]User)
		//dataMap = dataMap.(map[string]User)
	}
	for _, user := range data.Items {
		user.Name = user.First_name + " " + user.Last_name
		if strings.Contains(user.Name, query) || strings.Contains(user.About, query) {
			switch order_field {
			case fields.id:
				dataMap[strconv.Itoa(user.Id)] = user
			case fields.age:
				dataMap[strconv.Itoa(user.Age)] = user
			case fields.name:
				dataMap[user.Name] = user
			}
		}
	}
	keys := make([]string, 0, len(dataMap))
	for k := range dataMap {
		// switch k := k.(type) {
		// case int:
		// 	keys = append(keys, strconv.Itoa(k))
		// case string:
		// 	keys = append(keys, k)
		// }
		keys = append(keys, k)

	}
	fmt.Println("keys", keys)
	fmt.Println()

	//sort.Strings(keys)
	if order_by != OrderByAsIs {
		// slices.SortFunc()
		// slices.SortStableFunc(keys, func(a, b string) int {
		// 	switch order_by {
		// 	case OrderByAsc:
		// 		return keys[i] <= keys[j] // Сортировка в прямом порядке
		// 	case OrderByDesc:
		// 		return keys[i] >= keys[j] // Сортировка в обратном порядке
		// }
		sort.Slice(keys, func(i, j int) bool {
			switch order_by {
			case OrderByDesc:
				return keys[i] <= keys[j] // Сортировка в обратном порядке
			default: // OrderByAsc
				return keys[i] >= keys[j] // Сортировка в прямом порядке
			}
		})
	}
	fmt.Println("sorted_keys", keys)
	fmt.Println()
	// срез отсортированных данных
	outputData := make([]User, len(keys))
	for idx, key := range keys {
		some := dataMap[key]
		outputData[idx] = some
	}

	// результат - срез данных
	if offset > len(outputData) {
		outputData = []User{}
	} else if offset+limit > len(outputData) {
		outputData = outputData[offset:]
	} else {
		outputData = outputData[offset : offset+limit]
	}
	for _, user := range outputData {
		fmt.Println(user.Id, user.Name, user.Gender)
	}
	fmt.Println()
	// ответ в виде JSON
	result, err := json.Marshal(outputData)
	if err != nil {
		fmt.Println("Ошибка JSON ", err)
		http.Error(w, fmt.Sprintln("Ошибка JSON ", err), http.StatusInternalServerError)
		return
	}
	w.Write(result)

}

func main() {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	fmt.Println("start", ts.URL)
	fmt.Scanln()
	ts.Close()
	fmt.Println("stop")
}
