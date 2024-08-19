import heapq


# Структура для хранения информации о счетах
class Account:
    def __init__(self):
        self.saldo = 0.0
        self.position = 0
        self.turnover = 0.0
        self.trade_amount = 0


# Модуль обработки заявок
def process_orders(orders: list):
    # Очереди для лимитных заявок
    buy_orders = []
    sell_orders = []
    
    # Словарь для хранения данных по аккаунтам
    accounts = {}

    # Обработка каждой заявки
    for order in orders:
        order_id = order["order_id"]
        order_type = order["type"]
        account_id = order["account_id"]
        direction = order["dir"]
        price = order["price"]
        amount = order["amount"]
        
        if account_id not in accounts:
            accounts[account_id] = Account()
        
        # Обработка рыночных заявок
        if order_type == "market":
            if direction == 0:  # Покупка (BID)
                execute_market_order(sell_orders, accounts, account_id, price, amount, is_buy=True)
            else:  # Продажа (ASK)
                execute_market_order(buy_orders, accounts, account_id, price, amount, is_buy=False)
        
        # Обработка лимитных заявок
        elif order_type == "limit":
            if direction == 0:  # Покупка (BID)
                heapq.heappush(buy_orders, (-price, order_id, amount, account_id))
                execute_limit_order(buy_orders, sell_orders, accounts)
            else:  # Продажа (ASK)
                heapq.heappush(sell_orders, (price, order_id, amount, account_id))
                execute_limit_order(sell_orders, buy_orders, accounts)
        
        # Обработка специфических типов заявок (IOC, FOK)
        elif order_type == "ioc":
            breakpoint()
            if direction == 0:  # Покупка (BID)
                execute_ioc_order(sell_orders, accounts, account_id, price, amount, is_buy=True)
            else:  # Продажа (ASK)
                execute_ioc_order(buy_orders, accounts, account_id, price, amount, is_buy=False)

        elif order_type == "fok":
            breakpoint()
            if direction == 0:  # Покупка (BID)
                execute_fok_order(sell_orders, accounts, account_id, price, amount, is_buy=True)
            else:  # Продажа (ASK)
                execute_fok_order(buy_orders, accounts, account_id, price, amount, is_buy=False)

    return accounts


# Функция исполнения рыночной заявки
def execute_market_order(
    counter_orders: list,
    accounts: dict,
    account_id: int,
    price: float,
    amount: int, 
    is_buy: bool,
):
    while amount > 0 and counter_orders:
        best_price, _, best_amount, best_account_id = counter_orders[0]
        if (is_buy and price >= best_price) or (not is_buy and price <= abs(best_price)):
            trade_amount = min(amount, best_amount)
            accounts[account_id].trade_amount += trade_amount
            accounts[best_account_id].trade_amount += trade_amount

            trade_value = trade_amount * abs(best_price)
            accounts[account_id].turnover += trade_value
            accounts[best_account_id].turnover += trade_value
            
            if is_buy:
                accounts[account_id].position += trade_amount
                accounts[account_id].saldo -= trade_value
                accounts[best_account_id].position -= trade_amount
                accounts[best_account_id].saldo += trade_value
            else:
                accounts[account_id].position -= trade_amount
                accounts[account_id].saldo += trade_value
                accounts[best_account_id].position += trade_amount
                accounts[best_account_id].saldo -= trade_value

            amount -= trade_amount
            if trade_amount == best_amount:
                heapq.heappop(counter_orders)
            else:
                counter_orders[0] = (best_price, _, best_amount - trade_amount, best_account_id)
                heapq.heapify(counter_orders)
                counter_orders.sort()
            print(f"Market order on: {_}")
        else:
            break


# Функция исполнения лимитной заявки
def execute_limit_order(
    own_orders: list,
    counter_orders: list,
    accounts: dict,
):
    while own_orders and counter_orders:
        best_own_price, best_own_order_id, own_amount, own_account_id = own_orders[0]
        best_counter_price, best_counter_order_id, counter_amount, counter_account_id = counter_orders[0]
        
        if best_own_price >= best_counter_price:
            trade_amount = min(own_amount, counter_amount)
            accounts[own_account_id].trade_amount += trade_amount
            accounts[counter_account_id].trade_amount += trade_amount

            trade_value = trade_amount * best_counter_price
            accounts[own_account_id].turnover += trade_value
            accounts[counter_account_id].turnover += trade_value

            accounts[own_account_id].position += trade_amount
            accounts[counter_account_id].position -= trade_amount
            
            accounts[own_account_id].saldo -= trade_value
            accounts[counter_account_id].saldo += trade_value

            if trade_amount == own_amount:
                heapq.heappop(own_orders)
            else:
                own_orders[0] = (best_own_price, best_own_order_id, own_amount - trade_amount, own_account_id)
                heapq.heapify(own_orders)
                own_orders.sort()
            
            if trade_amount == counter_amount:
                heapq.heappop(counter_orders)
            else:
                counter_orders[0] = (best_counter_price, best_counter_order_id, counter_amount - trade_amount, counter_account_id)
                heapq.heapify(counter_orders)
                counter_orders.sort()
            print(f"Limit order on: {best_own_order_id}")
            print(f"Limit order on: {best_counter_order_id}")
        else:
            break


# Функция исполнения IOC заявки
def execute_ioc_order(
    counter_orders: list,
    accounts: dict,
    account_id: int,
    price: float,
    amount: int, 
    is_buy: bool,
):
    initial_amount = amount
    if counter_orders:
        best_price, _, best_amount, best_account_id = counter_orders[0]
        if (is_buy and price >= best_price) or (not is_buy and price <= abs(best_price)):
            trade_amount = min(amount, best_amount)
            accounts[account_id].trade_amount += trade_amount
            accounts[best_account_id].trade_amount += trade_amount

            trade_value = trade_amount * abs(best_price)
            accounts[account_id].turnover += trade_value
            accounts[best_account_id].turnover += trade_value
            
            if is_buy:
                accounts[account_id].position += trade_amount
                accounts[account_id].saldo -= trade_value
                accounts[best_account_id].position -= trade_amount
                accounts[best_account_id].saldo += trade_value
            else:
                accounts[account_id].position -= trade_amount
                accounts[account_id].saldo += trade_value
                accounts[best_account_id].position += trade_amount
                accounts[best_account_id].saldo -= trade_value

            amount -= trade_amount
            if trade_amount == best_amount:
                heapq.heappop(counter_orders)
            else:
                counter_orders[0] = (best_price, _, best_amount - trade_amount, best_account_id)
                heapq.heapify(counter_orders)
                counter_orders.sort()
            print(f"IOK order on: {_}")

    if amount == initial_amount:
        # Если IOC заявка не исполнилась, она отменяется
        return


# Функция исполнения FOK заявки
def execute_fok_order(
    counter_orders: list,
    accounts: dict,
    account_id: int,
    price: float,
    amount: int, 
    is_buy: bool,
):
    can_fulfill = False
    total_trade_amount = 0
    total_trade_value = 0

    if counter_orders:
        best_price, _, best_amount, best_account_id = counter_orders[0]
        print(f"FOK order on: {_}")
        if (is_buy and price >= best_price) or (not is_buy and price <= abs(best_price)):
            if best_amount >= amount:
                can_fulfill = True
                total_trade_amount = amount
                total_trade_value = total_trade_amount * abs(best_price)

    if can_fulfill:
        accounts[account_id].trade_amount += total_trade_amount
        accounts[best_account_id].trade_amount += total_trade_amount

        accounts[account_id].turnover += total_trade_value
        accounts[best_account_id].turnover += total_trade_value

        if is_buy:
            accounts[account_id].position += total_trade_amount
            accounts[account_id].saldo -= total_trade_value
            accounts[best_account_id].position -= total_trade_amount
            accounts[best_account_id].saldo += total_trade_value
        else:
            accounts[account_id].position -= total_trade_amount
            accounts[account_id].saldo += total_trade_value
            accounts[best_account_id].position += total_trade_amount
            accounts[best_account_id].saldo -= total_trade_value

        heapq.heappop(counter_orders)
    else:
        # Если FOK заявка не может быть полностью исполнена, она отменяется
        return
