def calculate_eth_for_wallets(initial_eth, initial_tokens, num_wallets, target_percentage):
    x = initial_eth
    y = initial_tokens
    results = []
    total_eth = 0
    delta_y = (target_percentage / 100) * initial_tokens / num_wallets

    for _ in range(num_wallets):
        delta_x = (delta_y * x) / (y - delta_y)
        results.append(delta_x)
        total_eth += delta_x

        x += delta_x
        y -= delta_y

    return results, total_eth

initial_eth = 2.0
initial_tokens = 100 
num_wallets = 50 
target_percentage = 80

required_eth, total_required_eth = calculate_eth_for_wallets(initial_eth, initial_tokens, num_wallets, target_percentage)

for index, eth in enumerate(required_eth, 1):
    print(f"Wallet {index}: Needs {eth:.4f} ETH")

print(f"Total ETH required by all wallets: {total_required_eth:.4f} ETH")