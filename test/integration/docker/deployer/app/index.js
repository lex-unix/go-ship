const http = require("http");
const fs = require("fs");

const server = http.createServer((_req, res) => {
    fs.readFile("version.txt", "utf8", (err, data) => {
        if (err) {
            console.error("Error reading version.txt:", err);
            res.statusCode = 500;
            res.setHeader("Content-Type", "text/plain");
            res.end("Internal Server Error\n");
        } else {
            res.statusCode = 200;
            res.setHeader("Content-Type", "text/plain");
            res.end(data.trim() + "\n");
        }
    });
});

const port = 3000;
server.listen(port, () => {
    console.log(`Server running at http://localhost:${port}/`);
});
