const express = require("express");
const http = require("http");
const WebSocket = require("ws");
const path = require("path");

const app = express();
const server = http.createServer(app);
const wss = new WebSocket.Server({ server });

require('dotenv').config();

const ADDR = process.env.FRONTEND_ADDR || "localhost";
const BIND_ADDR = process.env.FRONTEND_BIND_ADDR || "0.0.0.0";
const PORT = process.env.FRONTEND_PORT || 3000;
const BACKEND_ADDR = process.env.BACKEND_ADDR || "localhost";
const BACKEND_PORT = process.env.BACKEND_PORT || 8766;

// 静态文件服务
app.use(express.static(path.join(__dirname, "public")));

// 为每个页面添加路由
app.get("/", (req, res) => {
  res.sendFile(path.join(__dirname, "public/pages/index.html"));
});

app.get("/about", (req, res) => {
  res.sendFile(path.join(__dirname, "public/pages/about.html"));
});



const files = require("fs");
app.get('/api', (require, response) => {
  const query = require.query;
  if (query.type == "ai") {};
  if (query.type == "files") {
    if (query.command == "read") {
      files.stat(query.path, (error, stats) => {
        if (error) {
          if (error.code == 'ENOENT') {
            console.error('没有文件');
          } else {
            console.error(error);
          }
          return;
        };
        if (stats.isFile()) {
          files.readFile(query.path, "utf8", (error, data) => {
            if (error) {
              console.error(error);
              return;
            }
            response.json({ message: data });
          });
        };
        if (stats.isDirectory()) {
          files.readdir(query.path, (error, files) => {
            if (error) {
              console.error(error);
              return;
            }
            response.json({ message: files });
          });
        };
      });
    };
    if (query.command == "check") {
      files.stat(query.path, (error, stats) => {
        if (error) {
          response.json({ message: "没有文件" });
        } else {
          response.json({ message: "存在文件" });
        };
      });
    };
    if (query.command == "create") {
      files.mkdir(query.path, (error) => {
        if (error) {
          console.error(error);
          return;
        };
        response.json({ message: "创建成功" });
      });
    }
    if (query.command == "delete") {
      files.stat(query.path, (error, stats) => {
        if (error) {
          if (error.code == 'ENOENT') {
            console.error('没有文件');
          } else {
            console.error(error);
          };
          return;
        };
        if (stats.isFile()) {
          files.unlink(query.path, (error) => {
            if (error) {
              console.error(error);
              return;
            };
            response.json({ message: "删除成功" });
          });
        };
        if (stats.isDirectory()) {
          files.rmdir(query.path, { recursive: true }, (error) => {
            if (error) {
              console.error(error);
              return;
            };
            response.json({ message: "删除成功" });
          });
        };
      });
    };
    if (query.command == "write") {
      files.writeFile(query.path, query.data, (error) => {
        if (error) {
          console.error(error);
          return;
        };
        response.json({ message: "写入成功" });
      });
    };
    if (query.command == "rename") {
      files.rename(query.path, path.join(path.dirname(query.path), query.name), (error) => {
        if (error) {
          console.error(error);
          return;
        };
        response.json({ data: "命名成功" });
      });
    };
    if (query.command == "move") {
      files.rename(query.source, path.join(query.target, path.basename(query.source)), (error) => {
        if (error) {
          console.error(error);
          return;
        };
        response.json({ data: "移动成功" });
      });
    };
    if (query.command == "copy") {
      files.copyFile(query.source, path.join(query.target, path.basename(query.source)), (error) => {
        if (error) {
          console.error(error);
          return;
        };
        response.json({ data: "复制成功" });
      });
    };
  };
});




// 连接到 Python WebSocket 服务
let pythonSocket = null;

function connectToPython() {
  // 添加协议头
  const ws = new WebSocket(`ws://${BACKEND_ADDR}:${BACKEND_PORT}`, {
    headers: {
      Origin: `http://${ADDR}:${PORT}`,
      Connection: "Upgrade",
      Upgrade: "websocket",
    },
  });

  // 添加握手验证
  ws.onopen = () => {
    console.log("WebSocket 握手成功");
    ws.send(
      JSON.stringify({
        type: "handshake",
        protocol: "websocket",
        version: "13", // WebSocket 协议版本
      })
    );
  };

  ws.on("open", () => {
    console.log("已连接到 Python 服务");
    pythonSocket = ws;
  });

  ws.on("message", (message) => {
    try {
      const data = JSON.parse(message);

      // 添加音频URL路径
      if (data.audioFile) {
        data.audioUrl = `/audio/${data.audioFile}`;
      }

      // 直接转发，不修改顺序
      wss.clients.forEach((client) => {
        if (client.readyState === WebSocket.OPEN) {
          client.send(JSON.stringify(data));
        }
      });
    } catch (e) {
      console.error("处理Python响应出错:", e);
    }
  });

  ws.on("error", (error) => {
    console.error("Python 服务连接错误:", error);
    pythonSocket = null;
  });

  ws.on("close", () => {
    console.log("Python 服务连接断开，5秒后重试...");
    pythonSocket = null;
    setTimeout(connectToPython, 5000);
  });

  return ws;
}

// 初始连接
connectToPython();

// WebSocket 连接处理
wss.on("connection", (ws) => {
  console.log("新的客户端连接");

  ws.on("message", (message) => {
    console.log("收到客户端消息:", message);

    // 将消息转发到 Python 服务
    if (pythonSocket && pythonSocket.readyState === WebSocket.OPEN) {
      pythonSocket.send(message);
    } else {
      ws.send(
        JSON.stringify({
          type: "error",
          message: "暂时无法连接到后端服务",
        })
      );
    }
  });

  ws.on("close", () => {
    console.log("客户端断开连接");
  });
});

// 启动服务器
server.listen(PORT, BIND_ADDR, () => {
  console.log(`服务器运行在 http://${BIND_ADDR}:${PORT}`);
});

