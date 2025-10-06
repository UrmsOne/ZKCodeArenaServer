编译请求:
{
  "cmd": [{
    "args": ["/usr/bin/javac", "-encoding", "UTF-8", "Main.java"],
    "env": [
      "PATH=/usr/bin:/bin",
      "JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64"
    ],
    "files": [
      {"content": ""},
      {"name": "stdout", "max": 10240},
      {"name": "stderr", "max": 10240}
    ],
    "cpuLimit": 10000000000,
    "memoryLimit": 268435456,
    "procLimit": 50,
    "copyIn": {
      "Main.java": {
        "content": "import java.util.*;\n\npublic class Main {\n    public static void main(String[] args) {\n        Scanner scanner = new Scanner(System.in);\n        int s = scanner.nextInt();\n        int n = scanner.nextInt();\n        int d = scanner.nextInt();\n        \n        int[] w = new int[d + 1];\n        int[] v = new int[d + 1];\n        \n        for (int i = 1; i <= d; i++) {\n            w[i] = scanner.nextInt();\n            v[i] = scanner.nextInt();\n        }\n        \n        for (int year = 1; year <= n; year++) {\n            int m = s / 1000;\n            int[] dp = new int[m + 1];\n            \n            for (int i = 1; i <= d; i++) {\n                int bondUnits = w[i] / 1000;\n                int bondProfit = v[i];\n                \n                for (int j = bondUnits; j <= m; j++) {\n                    if (j >= bondUnits) {\n                        dp[j] = Math.max(dp[j], dp[j - bondUnits] + bondProfit);\n                    }\n                }\n            }\n            \n            s += dp[m];\n        }\n        \n        System.out.println(s);\n        scanner.close();\n    }\n}"
      }
    },
    "copyOut": ["stdout", "stderr"],
    "copyOutCached": ["Main.class"]
  }]
}


响应:[
    {
        "status": "Accepted",
        "exitStatus": 0,
        "time": 1025078000,
        "memory": 58527744,
        "runTime": 374842841,
        "procPeak": 31,
        "files": {
            "stderr": "",
            "stdout": ""
        },
        "fileIds": {
            "Main.class": "PUVGCWHP"
        }
    }
]

运行
{
  "cmd": [{
    "args": ["/usr/bin/java", "-server", "-XX:+UseSerialGC", "-XX:MaxRAMPercentage=50", "-XX:InitialRAMPercentage=25", "Main"],
    "env": [
      "PATH=/usr/bin:/bin",
      "JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64",
      "_JAVA_OPTIONS=-Djava.awt.headless=true"
    ],
    "files": [
      {"content": "10000\n4\n2\n4000\n400\n3000\n250"},
      {"name": "stdout", "max": 10240},
      {"name": "stderr", "max": 10240}
    ],
    "cpuLimit": 1000000000,
    "memoryLimit": 268435456,
    "procLimit": 100,
    "copyIn": {
      "Main.class": {
        "fileId": "PUVGCWHP"
      }
    },
    "copyOut": ["stdout", "stderr"]
  }]
}

[
    {
        "status": "Accepted",
        "exitStatus": 0,
        "time": 123554000,
        "memory": 24797184,
        "runTime": 75088572,
        "procPeak": 15,
        "files": {
            "stderr": "Picked up _JAVA_OPTIONS: -Djava.awt.headless=true\n",
            "stdout": "14050\n"
        }
    }
]