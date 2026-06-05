# Workspace Example Usage

### 1. Navigate to the *Workspaces* page via the page menu
![page menu](img/workspace_example_page_menu.webp)

### 2. Create a new workspace
Give the new workspace a name and choose the blockchain the workspace should operate on.
![new workspace](img/workspace_example_new_workspace.webp)

### 3. Navigate to the newly created workspace
Click on the workspace card to navigate to the workspace editor. The list also shows the workspace's blockchain.
![workspace overview](img/workspace_example_workspace_overview.webp)

### 4. Add a transaction to the workspace
In the workspace editor, click on *Add Entities* and copy and paste transaction hash `16c3ba4f36a704f400f682e0e7d21df78a93d10ce444f43fcfc2145a8f9a3d93` into the search field of the dialog. Hit enter or click on *Add Entity* to add the transaction to the workspace. A node, representing the transaction, will appear shortly after in the workspace. Click on the transaction node to view transaction details. 

The *Add Entities* dialog also allows to add multiple nodes at once. This can be achieved either by copy and pasting text containing multiple transaction hashes or addresses in the search bar, or by uploading a file containing multiple hashes.
![add transaction](img/workspace_example_add_transaction.webp)

### 5. Open the CoinJoin heuristic sidebar
After clicking on the newly added transaction, the transaction sidebar opens. Here  transaction details can be viewed. The top section of the sidebar allows various actions, including creating a CoinJoin heuristic. Clicking *Add CoinJoin Heuristic* opens the CoinJoin heuristic sidebar.
![open coinjoin heuristic sidebar](img/workspace_example_open_coinjoin_heuristic_side_bar.webp)

### 6. Create a CoinJoin heuristic
Via this sidebar CoinJoin heuristics can be created. Certain heuristic types are only compatible with certain types of parent nodes. In this case we want to select *Reverse lookup by time* and choose a maximum duration of 12 hours. After clicking *Add* the heuristic will be added to the workspace.

Depending on the chosen parameters, the execution of the heuristic might take a while. Usually it should only take a few seconds. After the heuristic has finished processing, it will show the number of found clusters on its node.

![coinjoin heuristic creation](img/workspace_example_coinjoin_heuristic_creation.webp)

### 7. Combine two CoinJoin heuristics
After the newly created heuristic is finished executing, clicking on it displays the heuristic details and its results. As its type is *Reverse Lookup* it will include [origin transactions](dash/originTransaction.md) which are potentially responsible for funding the [destination transaction](wasabi/destinationTransaction.md) `16c...` . In this case the heuristic returned roughly 1 900 clusters. 

To reduce this number, another heuristic can be applied on top. The new heuristic will use the results of the initial heuristic. In the heuristic details sidebar, click *Add CoinJoin Heuristic* and add the *Reverse amount* heuristic. This should result in a graph like shown below. The *Reverse amount* heuristic limits the results to clusters which have more or equal funds as the inputs of the destination transaction. As the destination transaction spends over 10 BTC, this is a strong filter and brings the number of clusters down to 44. 

Note: Due to ongoing address clustering the number found transactions or address clusters might change over time.

![reverse amount heuristic](img/workspace_example_reverse_amount.webp)

### 8. Extract CoinJoin heuristic results
Click on the *Reverse amount* heuristic to show the heuristic details' sidebar. Select two the transactions in the table and click on *Add Entities*. The selected transactions will appear in the workspace afterward. In this example two transactions, that belong to the same cluster (see cluster index in table), have been chosen. 

Because both of the transactions are results of both of the heuristics, connections between them are shown in the workspace editor.

![coinjoin heuristic details](img/workspace_example_coinjoin_heuristic_details.webp)

### 9. Add an address cluster
Click on one of the added transactions to open the transaction sidebar. Locate an input address and click on it. A dialog will appear which allows choosing between opening the address page or adding the address to the workspace. Choose *Add to Workspace*.
![add address](img/workspace_example_add_address.webp)

## 10. View node connections
Connections between nodes can carry important information. For example a connection between two address cluster reveals which transactions connect the clusters. Additionally, a connection between an address cluster and a CoinJoin heuristic shows which transactions connect the cluster to the heuristic.

## 11. Explore

Dakar allows viewing basic blockchain data, like transactions and addresses but also provides insights by analyzing CoinJoin graphs. Further explore features of Dakar by using workspaces, or discover more of them in the wiki.
