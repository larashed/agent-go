package tests

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func testClient() {
	payload := `{
	  "env": "staging",
	  "queries": [
		{
		  "created_at": "2018-05-29T14:23:00+00:00",
		  "query": "select * from 
		users
		 where 
		id
		 = ? limit 1",
		  "connection": "mysql",
		  "processed_in": 2
		},
		{
		  "created_at": "2018-05-29T14:23:00+00:00",
		  "query": "select * from 
		accounts
		 where 
		accounts
		.
		id
		 = ? limit 1",
		  "connection": "mysql",
		  "processed_in": 1
		}
	  ],
	  "job": [],
	  "request": {
		"created_at": "2018-05-29T14:22:58+00:00",
		"processed_in": 3936,
		"url": "http:\/\/larashed.local\/",
		"method": "GET",
		"route": {
		  "uri": "\/",
		  "name": "dashboard",
		  "action": "App\\Http\\Controllers\\DashboardController@index"
		},
		"user": {
		  "id": 1,
		  "name": "Ignas"
		},
		"meta": {
		  "referrer": null,
		  "user-agent": "Mozilla\/5.0 (Macintosh; Intel Mac OS X 10_13_3) AppleWebKit\/537.36 (KHTML, like Gecko) Chrome\/66.0.3359.181 Safari\/537.36",
		  "ip": "172.19.0.1"
		},
		"response": {
		  "code": 200,
		  "exception": [
			  {
				  "class": "Symfony\\Component\\HttpKernel\\Exception\\NotFoundHttpException",
				  "message": "Method call() does not exist",
				  "code": 0,
				  "file": "\/var\/www\/larashed\/backend\/vendor\/laravel\/framework\/src\/Illuminate\/Routing\/RouteCollection.php",
				  "line": 179,
				  "trace": [
					  {
						  "file": "\/var\/www\/larashed\/backend\/vendor\/laravel\/framework\/src\/Illuminate\/Routing\/Router.php",
						  "line": 612,
						  "function": "match",
						  "class": "Illuminate\\Routing\\RouteCollection"
					  },
					  {
						  "file": "\/var\/www\/larashed\/backend\/vendor\/laravel\/framework\/src\/Illuminate\/Routing\/Router.php",
						  "line": 601,
						  "function": "findRoute",
						  "class": "Illuminate\\Routing\\Router"
					  },
					  {
						  "file": "\/var\/www\/larashed\/backend\/vendor\/laravel\/framework\/src\/Illuminate\/Routing\/Router.php",
						  "line": 590,
						  "function": "dispatchToRoute",
						  "class": "Illuminate\\Routing\\Router"
					  },
					  {
						  "file": "\/var\/www\/larashed\/backend\/vendor\/laravel\/framework\/src\/Illuminate\/Foundation\/Http\/Kernel.php",
						  "line": 176,
						  "function": "dispatch",
						  "class": "Illuminate\\Routing\\Router"
					  }
				  ]
			  }
		  ]
		}
	  },
	  "webhook": []
	}
	`

	for i := 0; i < 50000; i++ {
		conn, err := net.Dial("tcp", "127.0.0.1:33101")
		if err != nil {
			fmt.Printf("conn %d\n", i)

			panic(err)
		}
		_, err = conn.Write([]byte(payload + "\n"))

		if err != nil {
			fmt.Printf("conn %d\n", i)

			panic(err)
		}

		conn.Close()

		if i%75 == 0 {
			fmt.Println("sleeping")
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func TestClient(t *testing.T) {
	testClient()
}
