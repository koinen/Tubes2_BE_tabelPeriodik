# Tabel Periodik
Tabel Periodik adalah program untuk mencari resep sebuah elemen pada gim Little Alchemy 2. Aplikasi ini merupakan penerapan algoritma BFS, DFS, serta Bidirectional. Algoritma ini akan mengasilkan sebuah tree yang setiap daunnya merupakan elemen dengan Tier 0. Repositori ini adalah backend dari aplikasi Tabel Periodik.

Algoritma yang Digunakan
1.	Breadth-First Search (BFS)
BFS mengeksplorasi semua kombinasi dari level paling awal terlebih dahulu. BFS menawarkan jalur dengan kedalaman yang rendah. 
2.	Depth-First Search (DFS)
DFS menelusuri lebih dalam ke satu jalur sebelum berpindah. DFS lebih cepat dalam beberapa kasus tetapi tidak menjamin jalur terpendek.
3.	Bidirectional Search
Algoritma ini mencari dari dua arah: dari elemen dasar ke target, dan dari target ke elemen dasar, bertemu di tengah. Hal ini mempercepat pencarian secara signifikan.

# Dependency
Hal yang perlu di unduh sebelum menggunakan program ini adalah
1. Docker

# Cara Menjalankan
🔧 1. Jalankan dengan Docker
bash
Copy
Edit
docker build -t alchemy-backend .
docker run -p 8080:8080 alchemy-backend
Tabel Periodik
Tabel Periodik adalah program untuk mencari resep sebuah elemen pada gim Little Alchemy 2. Aplikasi ini merupakan penerapan algoritma BFS, DFS, serta Bidirectional. Algoritma ini akan mengasilkan sebuah tree yang setiap daunnya merupakan elemen dengan Tier 0. Repositori ini adalah backend dari aplikasi Tabel Periodik.

Algoritma yang Digunakan
1.	Breadth-First Search (BFS)
BFS mengeksplorasi semua kombinasi dari level paling awal terlebih dahulu. BFS menawarkan jalur dengan kedalaman yang rendah. 
2.	Depth-First Search (DFS)
DFS menelusuri lebih dalam ke satu jalur sebelum berpindah. DFS lebih cepat dalam beberapa kasus tetapi tidak menjamin jalur terpendek.
3.	Bidirectional Search
Algoritma ini mencari dari dua arah: dari elemen dasar ke target, dan dari target ke elemen dasar, bertemu di tengah. Hal ini mempercepat pencarian secara signifikan.

Dependency
Hal yang perlu di unduh sebelum menggunakan program ini adalah
1.	Go 1.20 atau lebih baru
2.	Docker

Cara Menjalankan
🔧 1. Jalankan dengan Docker
bash
Copy
Edit
docker build -t alchemy-backend .
docker run -p 8080:8080 alchemy-backend
