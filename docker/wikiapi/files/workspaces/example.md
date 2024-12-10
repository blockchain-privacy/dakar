# Workspace Example Usage

### 1. Navigate to the *Workspaces* page via the page menu
![page menu](img/workspace_example_page_menu.webp)

### 2. Create a new workspace
![new workspace](img/workspace_example_new_workspace.webp)

### 3. Navigate to the newly created workspace
![workspace overview](img/workspace_example_workspace_overview.webp)

### 4. Add a transaction in the workspace
Copy and paste transaction hash `85834761b7befb9852e4fd328f316ff6f190d945dfc8ee8e770bbe69695afd09` into the workspace search bar and hit enter to add the transaction to the workspace. The transaction will appear shortly after in the workspace. Click on the the transaction node to view transaction details.
![add transaction](img/workspace_example_add_transaction.webp)

### 5. Open the CoinJoin heuristic side bar
After clicking on the newly added transaction, the transaction side bar opens. Here  transaction details can be viewed. The top section of the sidbar allows various actions, including creating a CoinJoin heuristic. Clicking *Add CoinJoin Heuristic* opens the CoinJoin heuristic side bar.
![open coinjoin heuristic sidebar](img/workspace_example_open_coinjoin_heuristic_side_bar.webp)

### 6. Create a CoinJoin heuristic
Via this side bar CoinJoin heuristics can be created. Certain heuristic types are only compatible with certain types of parent nodes. In this case we want to select *Reverse Lookup* and choose a lookback time of 12 hours. After clicking *Add* the heuristic will be added to the workspace.
![coinjoin heuristic creation](img/workspace_example_coinjoin_heuristic_creation.webp)

### 7. Extract CoinJoin heuristic results
After the newly created heuristic is finished executing, clicking on it displays the heuristic details and its results. As its type is *Reverse Lookup* it will include [origin transactions](dash/originTransaction.md) which are potentially responsible for funding the [destination transaction](dash/destinationTransaction.md) `858...` .

Select transaction `2f9c31e2a987086fc9fdc3a175769c09e0e6fe38c0dad29c5351d71f5d3ee2b1` and `4a4a3af8dc0d6955ca9ec34c0f8a3c3d33ed2b598a6b8604ae1e3f902d073a57` in the side bar and click *Add entities*. The selected transactions will appear in the workspace afterwards.
![coinjoin heuristic details](img/workspace_example_coinjoin_heuristic_details.webp)


### 8. Add an address cluster
Click on transaction `4a4...` to open its transaction side bar. Locate input address `Xmp4FTCTXw78ANi9Hjx4VwqHVgwDZxtsvY` and click on it. A dialog will appear which allows choosing between opening the address page or adding the address to the workspace. Choose *Add to Workspace*.
![add address](img/workspace_example_add_address.webp)

## 9. View node connections
Connections between nodes can carry important information. For example a connection between two address cluster reveals which transactions connect the clusters. Additionally, a connection between an address cluster and a CoinJoin heuristic shows which transactions connect the cluster to the heuristic.

Click on the arrow connecting cluster `Xmp...` from the previous step with the heuristic to show the connection side bar.
