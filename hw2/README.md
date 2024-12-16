Made 3 stress tests:
1. **100 requests** at 19:30 (**Concurrency 10**; time taken 0.174 seconds) 
2. **1000 requests** at 19:35 (**Concurrency 10**; time taken 0.890 seconds) 
3. **1000 requests** at 19:40 (**Concurrency 20**; time taken 1.471 seconds - higher concurrency doesn’t always lead to better results)

Elasticsearch metrics
![image](https://github.com/user-attachments/assets/aac09dcd-e83b-4b99-b705-3b2b380c9e46)
![image](https://github.com/user-attachments/assets/106f3f5a-6cba-45f8-99bd-e828af2d380e)

MongoDB metrics
![image](https://github.com/user-attachments/assets/c94f28b6-75cd-4316-8956-b9f8457bf5ee)
![image](https://github.com/user-attachments/assets/cb00de24-cb7c-493f-b5d6-495f80d67533)

