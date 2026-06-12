Feature: process user.created event

  Background: the barong-users-manager is up and running
    Given a barong login endpoint:
    """json
    {
      "id": "barongLoginSucceed",
      "httpRequest": {
        "method": "POST",
        "path": "/api/v1/auth/identity/sessions"
      },
      "httpResponse": {
        "statusCode": 200,
        "headers": {
          "content-type": ["application/json"],
          "Set-Cookie": ["_barong_session=test_session; Path=/; HttpOnly"]
        },
        "body": {
          "uid": "IDADMIN000001",
          "email": "admin@barong.io",
          "role": "admin",
          "level": 4,
          "otp": false,
          "state": "active"
        }
      },
      "priority": 0,
      "timeToLive": {"unlimited": true},
      "times": {"unlimited": true}
    }
    """
    And a running barong-users-manager

  Scenario: a user.created event adds an email verified label to the user
    Given a user.created event:
    """json
    {
      "payload": "eyJpc3MiOiJiYXJvbmciLCJqdGkiOiIzNTQxZWUxYy05NzUzLTRmOTgtOWEyOC1hMTQ1ODMyZTEyZWMiLCJpYXQiOjE3Nzg0NTc5MTUsImV4cCI6MTc3ODQ2MTUxNSwiZXZlbnQiOnsicmVjb3JkIjp7InVpZCI6IklERUMzMEM5RjU2NiIsImVtYWlsIjoidGhhZEByZWljaGVsLmluZm8iLCJyb2xlIjoibWVtYmVyIiwibGV2ZWwiOjAsIm90cCI6ZmFsc2UsInN0YXRlIjoiYWN0aXZlIiwiY3JlYXRlZF9hdCI6IjIwMjYtMDUtMTFUMDA6MDU6MTVaIiwidXBkYXRlZF9hdCI6IjIwMjYtMDUtMTFUMDA6MDU6MTVaIn0sIm5hbWUiOiJtb2RlbC51c2VyLmNyZWF0ZWQifX0",
      "signatures": [
        {
          "protected": "eyJhbGciOiJSUzI1NiJ9",
          "header": {"kid": "barong"},
          "signature": "dummysignature"
        }
      ]
    }
    """
    And a barong endpoint to add a label to a user:
    """json
    {
      "id": "addLabelSucceed",
      "httpRequest": {
        "method": "POST",
        "path": "/api/v1/auth/admin/users/labels"
      },
      "httpResponse": {
        "statusCode": 200,
        "headers": {"content-type": ["application/json"]},
        "body": {}
      },
      "priority": 0,
      "timeToLive": {"unlimited": true},
      "times": {"unlimited": true}
    }
    """
    When the event is published
    Then barong-users-manager adds the label to the user on the Barong Admin API
    And barong-users-manager produces the following log:
    """
    label added to user
    """
