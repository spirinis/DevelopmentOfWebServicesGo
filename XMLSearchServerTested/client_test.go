package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// код писать тут

func TestFindUsers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	client := SearchClient{
		AccessToken: "Отвечааааай!",
		URL:         ts.URL,
	}
	unauthorizedClient := SearchClient{
		AccessToken: "НеОтвечааааай!",
		URL:         ts.URL,
	}
	type TestCase struct {
		request SearchRequest
		Result  SearchResponse
		IsError bool
	}
	cases := []TestCase{
		{
			request: SearchRequest{
				Query:      "dolore",
				OrderField: "Id",
				OrderBy:    1,
				Limit:      10,
				Offset:     0,
			},
			Result: SearchResponse{
				Users:    []User{},
				NextPage: false,
			},
			IsError: false,
		},
		{
			request: SearchRequest{
				Query:      "dolore",
				OrderField: "Name",
				OrderBy:    -1,
				Limit:      10,
				Offset:     0,
			},
			Result: SearchResponse{
				Users:    []User{},
				NextPage: false,
			},
			IsError: false,
		},
		{
			request: SearchRequest{
				Query:      "Dickson Silva",
				OrderField: "",
				OrderBy:    0,
				Limit:      10,
				Offset:     0,
			},
			Result: SearchResponse{
				Users:    []User{},
				NextPage: false,
			},
			IsError: false,
		},
		{
			request: SearchRequest{
				Query:      "ipsum",
				OrderField: "Age",
				OrderBy:    0,
				Limit:      10,
				Offset:     0,
			},
			Result: SearchResponse{
				Users:    []User{},
				NextPage: false,
			},
			IsError: false,
		},
		{ // err limit < 0
			request: SearchRequest{
				Query:      "ТоЧегоНетВДанных",
				OrderField: "Id",
				OrderBy:    1,
				Limit:      -1,
				Offset:     0,
			},
		},
		{ // limit > 25
			request: SearchRequest{
				Query:      "ТоЧегоНетВДанных",
				OrderField: "Id",
				OrderBy:    1,
				Limit:      999,
				Offset:     0,
			},
		},
		{ // err Offset < 0
			request: SearchRequest{
				Query:      "ТоЧегоНетВДанных",
				OrderField: "Id",
				OrderBy:    1,
				Limit:      10,
				Offset:     -1,
			},
		},
	}

	// Без авторизации
	_, err := unauthorizedClient.FindUsers(
		SearchRequest{
			Query:      "dolore",
			OrderField: "Id",
			OrderBy:    1,
			Limit:      10,
			Offset:     0,
		})

	if err == nil {
		t.Errorf("[%d] expected error, got nil", 0)
	} else if errors.Is(err, fmt.Errorf("Bad AccessToken")) {
		t.Errorf("[%d] Ожидалась не эта ошибка: %#v", 0, err)
	}
	// с ошибкой
	type doError struct {
		Client   SearchClient
		resError error
	}
	errorClients := []doError{
		{
			SearchClient{
				AccessToken: "Уронись error timeout",
				URL:         ts.URL,
			},
			fmt.Errorf(""),
		},
		{
			SearchClient{
				AccessToken: "Уронись error other",
				URL:         ts.URL,
			},
			fmt.Errorf(""),
		},
	}
	for _, errorClient := range errorClients {
		errorClient.Client.FindUsers(
			SearchRequest{
				Query:      "ТоЧегоНетВДанных",
				OrderField: "Id",
				OrderBy:    1,
				Limit:      10,
				Offset:     0,
			})
	}

	for caseNum, item := range cases {
		_, err := client.FindUsers(item.request)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum+1, err)
		}
		if err == nil && item.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum+1)
		}
		// if !reflect.DeepEqual(item.Result, result) {
		// 	t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Result, result)
		// }
	}

	ts.Close()
}

/*
Требуется:
* Написать функцию SearchServer в файле `client_test.go`, который вы будете запускать в тесте через тестовый сервер (`httptest.NewServer`, пример использования в `4/http/server_test.go`)
* Покрыть тестами метод FindUsers, чтобы покрытие файла `client.go` было максимально возможным, а именно - 100%. Тесты писать в `client_test.go`. Но когда вы будете запускать тесты с флагом покрытия - там будет писаться общий процент, какой процент в `client.go` - смотрите в отчете.
* Так же требуется сгенерировать html-отчет с покрытием. См. пример построения тестового покрытия и отчета в `3/testing/coverage_test.go`.
* Тесты надо писать полноценные, те не чтобы получить покрытие, а которые реально тестируют ваш код, проверяют возвращаемый результат, граничные случаи и тд. Они должны показывать что SearchServer работает правильно.
* Из предыдущего пункта вытекает что SearchServer тоже надо писать полноценный


Дополнительно:
* Данные для работы лежаит в файле `dataset.xml`
* Как работать с XML - почти так же как с JSON, смотрите доку https://golang.org/pkg/encoding/xml/ и пример в боте
* Запускать как `go test -cover`
* Можно начать с того что вы просто напишите сервер в `main.go` который реализует логику, а потом уже унесите это в `client_test.go`
* Построение покрытия: `go test -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html`
* Документация https://golang.org/pkg/net/http/ может помочь
* Пользуйтесь табличным тестированием - это когда у вас есть слайс тест-кейсок, которые отличаются параметрами.
* Вы можете не ограничиваться функцией SearchServer при тестировании, если вам надо проверить какой-то совсем отдельный хитрый кейс, вроде ошибки. Но таких случаев будет немного. В основном всё будет в SearchServer
* Для покрытия тестом одной из ошибок придётся залезть в исходники функции, которая возвращает эту ошибку, и посмотреть при каких условиях работы или входных данных это происходит. Это клиентская ошибка, т.е. запрос в этом случае в сервер уходить не будет.
* Не пытайтесь реализовать таймаут подключением к неизвестному IPшнику
* Блок c NextPage на строке 121 в client.go используется для создания постраничной навигации - я заглядываю в сервер на +1 запись - если она есть - я могу показать следующую страницу

Объем кода:
* SearchServer со всеми структурами и всем-всем будет 170-200 строк
* Тесты 200-300-400 строк, в зависимости от формы - основное там будет список тест-кейсов

Рекомендуемый план работы:
1. Напишите в функции main код который просто по фиксированным параметрам реализует логику SearchServer и выводит в консоль, без http
2. Теперь оформите ваш код в http-хендлер, параметры уже не хардкодом, а берите из запроса
3. Проверьте запросами из браузера что код отрабатывает
4. Теперь начинайте писать тесты в client_test.go
5. Реализуйте сначала один тест, который просто делает запрос через SearchClient-а в ваш хттп-хендлер, запущенный через тестовый сервер
6. Теперь постройте отчет и смотрите какой код у вас был вызван, а какой нет
7. Начинайте дописывать тест кейсы
8. Для ошибок реализуйте отдельный хендлер или хендлеры
*/
