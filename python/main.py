import csv
import sys
from market_order_executor import process_orders


# Функция для чтения данных из CSV файла
def read_orders_from_csv(file_path):
    orders = []
    with open(file_path, mode='r') as file:
        reader = csv.DictReader(file)
        for row in reader:
            orders.append({
                "order_id": int(row["order_id"]),
                "type": row["type"],
                "account_id": int(row["account_id"]),
                "dir": int(row["dir"]),
                "price": float(row["price"]),
                "amount": int(row["amount"])
            })
    return orders


# Функция для записи результатов в CSV файл
def write_results_to_csv(file_path, accounts):
    with open(file_path, mode='w', newline='') as file:
        writer = csv.writer(file)
        writer.writerow(["account_id", "saldo", "position", "turnover", "trade_amount"])
        for account_id, account in accounts.items():
            writer.writerow([account_id, account.saldo, account.position, account.turnover, account.trade_amount])


def main(input_file, output_file):
    # Шаг 1: Чтение заявок из файла
    orders = read_orders_from_csv(input_file)
    
    # Шаг 2: Обработка заявок
    accounts = process_orders(orders)
    
    # Шаг 3: Запись результатов в файл
    write_results_to_csv(output_file, accounts)
    
    print(f"Processing completed. Results saved in {output_file}")


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python main.py <input_file> <output_file>")
        sys.exit(1)
    
    input_file = sys.argv[1]
    output_file = sys.argv[2]
    
    main(input_file, output_file)
