package main

import (
    "encoding/csv"
    "fmt"
    "log"
    "os"
    "strconv"
	"math"
	"container/heap"
)

// Структура для хранения информации о счетах
type Account struct {
    Saldo       float64
    Position    int
    Turnover    float64
    TradeAmount int
}

// Структура для хранения заявки
type Order struct {
	OrderID   int
	Type      string
	AccountID int
	Dir       int
	Price     float64
	Amount    int
}

// PriorityQueue реализует интерфейс heap.Interface и содержит Orders
type PriorityQueue []*Order

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
    return pq[i].Price > pq[j].Price // чем выше цена, тем выше приоритет
}

func (pq PriorityQueue) Swap(i, j int) {
    pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
    item := x.(*Order)
    *pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
    old := *pq
    n := len(old)
    item := old[n-1]
    *pq = old[0 : n-1]
    return item
}

// ProcessOrders обрабатывает заявки и возвращает информацию по аккаунтам
func ProcessOrders(orders []Order) map[int]*Account {
    buyOrders := &PriorityQueue{}
    sellOrders := &PriorityQueue{}
    heap.Init(buyOrders)
    heap.Init(sellOrders)

    accounts := make(map[int]*Account)

    for _, order := range orders {
        if _, ok := accounts[order.AccountID]; !ok {
            accounts[order.AccountID] = &Account{}
        }

        switch order.Type {
        case "market":
            if order.Dir == 0 { // Покупка (BID)
                executeMarketOrder(sellOrders, accounts, order.AccountID, order.Price, order.Amount, true)
            } else { // Продажа (ASK)
                executeMarketOrder(buyOrders, accounts, order.AccountID, order.Price, order.Amount, false)
            }

        case "limit":
            if order.Dir == 0 { // Покупка (BID)
                heap.Push(buyOrders, &Order{Price: -order.Price, OrderID: order.OrderID, Amount: order.Amount, AccountID: order.AccountID})
                executeLimitOrder(buyOrders, sellOrders, accounts)
            } else { // Продажа (ASK)
                heap.Push(sellOrders, &Order{Price: order.Price, OrderID: order.OrderID, Amount: order.Amount, AccountID: order.AccountID})
                executeLimitOrder(sellOrders, buyOrders, accounts)
            }

        case "ioc":
            if order.Dir == 0 { // Покупка (BID)
                executeIOCOrder(sellOrders, accounts, order.AccountID, order.Price, order.Amount, true)
            } else { // Продажа (ASK)
                executeIOCOrder(buyOrders, accounts, order.AccountID, order.Price, order.Amount, false)
            }

        case "fok":
            if order.Dir == 0 { // Покупка (BID)
                executeFOKOrder(sellOrders, accounts, order.AccountID, order.Price, order.Amount, true)
            } else { // Продажа (ASK)
                executeFOKOrder(buyOrders, accounts, order.AccountID, order.Price, order.Amount, false)
            }
        }
    }

    return accounts
}


func executeMarketOrder(counterOrders *PriorityQueue, accounts map[int]*Account, accountID int, price float64, amount int, isBuy bool) {
    for amount > 0 && counterOrders.Len() > 0 {
        bestOrder := heap.Pop(counterOrders).(*Order)
        if (isBuy && price >= bestOrder.Price) || (!isBuy && price <= math.Abs(bestOrder.Price)) {
            tradeAmount := min(amount, bestOrder.Amount)
            tradeValue := float64(tradeAmount) * math.Abs(bestOrder.Price)

            // Обновляем позиции и финансовые показатели для счетов
            accounts[accountID].TradeAmount += tradeAmount
            accounts[bestOrder.AccountID].TradeAmount += tradeAmount
            accounts[accountID].Turnover += tradeValue
            accounts[bestOrder.AccountID].Turnover += tradeValue

            if isBuy {
                accounts[accountID].Position += tradeAmount
                accounts[accountID].Saldo -= tradeValue
                accounts[bestOrder.AccountID].Position -= tradeAmount
                accounts[bestOrder.AccountID].Saldo += tradeValue
            } else {
                accounts[accountID].Position -= tradeAmount
                accounts[accountID].Saldo += tradeValue
                accounts[bestOrder.AccountID].Position += tradeAmount
                accounts[bestOrder.AccountID].Saldo -= tradeValue
            }

            // Уменьшаем количество оставшихся лотов для исполнения
            amount -= tradeAmount
            if tradeAmount < bestOrder.Amount {
                bestOrder.Amount -= tradeAmount
                heap.Push(counterOrders, bestOrder)
            }
        } else {
            heap.Push(counterOrders, bestOrder) // Заявка не подходит, возвращаем её в очередь
            break
        }
    }
}

func executeLimitOrder(ownOrders, counterOrders *PriorityQueue, accounts map[int]*Account) {
    for ownOrders.Len() > 0 && counterOrders.Len() > 0 {
        bestOwnOrder := heap.Pop(ownOrders).(*Order)
        bestCounterOrder := heap.Pop(counterOrders).(*Order)

        if bestOwnOrder.Price >= bestCounterOrder.Price {
            tradeAmount := min(bestOwnOrder.Amount, bestCounterOrder.Amount)
            tradeValue := float64(tradeAmount) * bestCounterOrder.Price

            // Обновляем позиции и финансовые показатели для счетов
            accounts[bestOwnOrder.AccountID].TradeAmount += tradeAmount
            accounts[bestCounterOrder.AccountID].TradeAmount += tradeAmount
            accounts[bestOwnOrder.AccountID].Turnover += tradeValue
            accounts[bestCounterOrder.AccountID].Turnover += tradeValue

            accounts[bestOwnOrder.AccountID].Position += tradeAmount
            accounts[bestCounterOrder.AccountID].Position -= tradeAmount
            accounts[bestOwnOrder.AccountID].Saldo -= tradeValue
            accounts[bestCounterOrder.AccountID].Saldo += tradeValue

            if tradeAmount < bestOwnOrder.Amount {
                bestOwnOrder.Amount -= tradeAmount
                heap.Push(ownOrders, bestOwnOrder)
            }
            if tradeAmount < bestCounterOrder.Amount {
                bestCounterOrder.Amount -= tradeAmount
                heap.Push(counterOrders, bestCounterOrder)
            }
        } else {
            // Заявки не совпадают по цене, возвращаем их в соответствующие очереди
            heap.Push(ownOrders, bestOwnOrder)
            heap.Push(counterOrders, bestCounterOrder)
            break
        }
    }
}

func executeIOCOrder(counterOrders *PriorityQueue, accounts map[int]*Account, accountID int, price float64, amount int, isBuy bool) {
    if counterOrders.Len() > 0 {
        bestOrder := heap.Pop(counterOrders).(*Order)
        if (isBuy && price >= bestOrder.Price) || (!isBuy && price <= math.Abs(bestOrder.Price)) {
            tradeAmount := min(amount, bestOrder.Amount)
            tradeValue := float64(tradeAmount) * math.Abs(bestOrder.Price)

            // Обновляем позиции и финансовые показатели для счетов
            accounts[accountID].TradeAmount += tradeAmount
            accounts[bestOrder.AccountID].TradeAmount += tradeAmount
            accounts[accountID].Turnover += tradeValue
            accounts[bestOrder.AccountID].Turnover += tradeValue

            if isBuy {
                accounts[accountID].Position += tradeAmount
                accounts[accountID].Saldo -= tradeValue
                accounts[bestOrder.AccountID].Position -= tradeAmount
                accounts[bestOrder.AccountID].Saldo += tradeValue
            } else {
                accounts[accountID].Position -= tradeAmount
                accounts[accountID].Saldo += tradeValue
                accounts[bestOrder.AccountID].Position += tradeAmount
                accounts[bestOrder.AccountID].Saldo -= tradeValue
            }

            if tradeAmount < bestOrder.Amount {
                bestOrder.Amount -= tradeAmount
                heap.Push(counterOrders, bestOrder)
            }
        } else {
            heap.Push(counterOrders, bestOrder) // Не исполнилось, возвращаем обратно
        }
    }
}

func executeFOKOrder(counterOrders *PriorityQueue, accounts map[int]*Account, accountID int, price float64, amount int, isBuy bool) {
    canFulfill := false
    totalTradeAmount := 0
    totalTradeValue := 0.0

    if counterOrders.Len() > 0 {
        bestOrder := heap.Pop(counterOrders).(*Order)
        if (isBuy && price >= bestOrder.Price && bestOrder.Amount >= amount) || (!isBuy && price <= math.Abs(bestOrder.Price) && bestOrder.Amount >= amount) {
            canFulfill = true
            totalTradeAmount = amount
            totalTradeValue = float64(totalTradeAmount) * math.Abs(bestOrder.Price)
        }

        if canFulfill {
            accounts[accountID].TradeAmount += totalTradeAmount
            accounts[bestOrder.AccountID].TradeAmount += totalTradeAmount
            accounts[accountID].Turnover += totalTradeValue
            accounts[bestOrder.AccountID].Turnover += totalTradeValue

            if isBuy {
                accounts[accountID].Position += totalTradeAmount
                accounts[accountID].Saldo -= totalTradeValue
                accounts[bestOrder.AccountID].Position -= totalTradeAmount
                accounts[bestOrder.AccountID].Saldo += totalTradeValue
            } else {
                accounts[accountID].Position -= totalTradeAmount
                accounts[accountID].Saldo += totalTradeValue
                accounts[bestOrder.AccountID].Position += totalTradeAmount
                accounts[bestOrder.AccountID].Saldo -= totalTradeValue
            }
        } else {
            // Если FOK заявка не может быть полностью исполнена, она отменяется
            heap.Push(counterOrders, bestOrder) // Заявка не исполнилась, возвращаем её обратно
        }
    }
}

// Функция для чтения данных из CSV файла
func readOrdersFromCSV(filePath string) ([]Order, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        return nil, err
    }

    var orders []Order
    for _, record := range records[1:] { // Пропускаем заголовок
        orderID, _ := strconv.Atoi(record[0])
        accountID, _ := strconv.Atoi(record[2])
        dir, _ := strconv.Atoi(record[3])
        price, _ := strconv.ParseFloat(record[4], 64)
        amount, _ := strconv.Atoi(record[5])

        orders = append(orders, Order{
            OrderID:   orderID,
            Type:      record[1],
            AccountID: accountID,
            Dir:       dir,
            Price:     price,
            Amount:    amount,
        })
    }
    return orders, nil
}

// Функция для записи результатов в CSV файл
func writeResultsToCSV(filePath string, accounts map[int]*Account) error {
    file, err := os.Create(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Запись заголовка
    writer.Write([]string{"account_id", "saldo", "position", "turnover", "trade_amount"})

    // Запись данных
    for accountID, account := range accounts {
        writer.Write([]string{
            strconv.Itoa(accountID),
            fmt.Sprintf("%.2f", account.Saldo),
            strconv.Itoa(account.Position),
            fmt.Sprintf("%.2f", account.Turnover),
            strconv.Itoa(account.TradeAmount),
        })
    }

    return nil
}

func main() {
    if len(os.Args) != 3 {
        fmt.Println("Usage: go run main.go <input_file> <output_file>")
        os.Exit(1)
    }

    inputFile := os.Args[1]
    outputFile := os.Args[2]

    // Шаг 1: Чтение заявок из файла
    orders, err := readOrdersFromCSV(inputFile)
    if err != nil {
        log.Fatalf("Error reading orders from CSV: %v", err)
    }

    // Шаг 2: Обработка заявок
    accounts := ProcessOrders(orders)

    // Шаг 3: Запись результатов в файл
    err = writeResultsToCSV(outputFile, accounts)
    if err != nil {
        log.Fatalf("Error writing results to CSV: %v", err)
    }

    fmt.Printf("Processing completed. Results saved in %s\n", outputFile)
}
