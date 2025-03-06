

____

### Observations from the Logarithmic Trendline:**


![alt text](./images/image1.png)


The logarithmic trendline confirms that **search**, **insert**, and **delete** operations in the AVL tree are performing close to `O(log(n))` complexity.

However, the execution spikes suggest additional overhead, likely due to:
- Balancing operations (especially for **insert** operation, spike could be happening due to a sequence of insertions that forces multiple rebalances).
- Memory/cache effects (may causing sudden hardware-related slowdowns).


____

### Observations from counting sort performance trends

![alt text](./images/image3.png)

All dataset sizes (`n`) show similar increasing trends as `k` grows  
This confirms that range (`k`) dominates the complexity, aligning with `O(n + k)`

**Key insights from the charts:**
- When `range` <= n, Counting sort performs well.
- When `range` > n, execution time increases rapidly.