# Используемые термины

goroutine - легкий поток, управляемый Go в рантайме

```go
go f(x, y, z)
```

channels - типизированый канал в который можно писать и читать из него

по дефолту чтение и запись - блокирующие операции, пока вторая сторона не будет готова. 
Такой подход позволяет синхронизировать горутины без явной блокировки.

каналы можно буферизовать: для этого нужно создать список из каналов

```go
bufferedChannel := make(chan struct{}, 10)
```

# Как выполнять асинхронный код в Golang

Предположим перед нами стоит задача: "нужно отправить N запросов по заданному url и получить данные"

Мы не хотим ждать выполнение последовательных запросов, и можем воспользоваться конкурентным выполнением неблокирующих запросов.

Напишем следующую функцию:

```go
type requestsData struct {
Url string
}

type responseData struct {
Body string
}

func asyncGetData(requestsData requestsData, responseChannel chan responseData, semaphoreChannel chan struct{}) {
	semaphoreChannel <- struct{}{}
	response, err := http.Get(requestsData.Url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}
	responseChannel <- responseData{
		Body: string(body),
	}
	<-semaphoreChannel
}
```

Она отправляет запрос по поданному урлу, парсим ответ и записывает результат в канал.
В первой строчке пишем в буферизованный канал пустую структуру 
в последней строчке читаем записанную структуру из него же -
этим действием мы будем контролировать количество работающих горутин.

Напишем код, который будет выполнять нашу функцию:
```go
func runCode(){
	url := "https://httpbin.org/ip"

	responseChannel := make(chan responseData)
	semaphoreChannel := make(chan struct{}, semaphoreChannelLength)
	requestsNumber := 1000
	for i := 0; i < requestsNumber; i++ {
		go asyncGetData(
			requestsData{
				Url: url,
			},
			responseChannel,
			semaphoreChannel,
		)
	}
	defer func() {
		close(responseChannel)
		close(semaphoreChannel)
	}()
	println("end requests")
	var results []responseData
	for {
		res := <-responseChannel
		results = append(results, res)
		if len(results) == requestsNumber {
			break
		}
	}
	println("end save response")
}
```

1. создаем канал с ответом
2. создаем буферизованный канал
3. выполняем 1000 get запросов по адресу `"https://httpbin.org/ip"`
4. явно закрываем каналы, показывая, что не будем в них писать
5. создаем массив для результатов
6. читаем из канала с ответом на запрос
7. записываем результат в массив
8. выходим из цикла, когда считали все ответы