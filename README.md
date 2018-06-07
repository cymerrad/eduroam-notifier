# Eduroam notifier

## Założenia

Następujący input ma przerobić na spam dla użytkownika. Spamujemy tylko raz z dokładnością do MAC-adresu urządzenia. I jakoś tak raz na tydzień.
Wszystko ofc konfigurowalne, najlepiej z użyciem klienta.

```json
{
  "check_result": {
    "result_description": "Stream received messages matching <action:\"Login incorrect (mschap: MS-CHAP2-Response is incorrect)\"> (Current grace time: 0 minutes)",
    "triggered_condition": {
      "id": "9242a930-183a-4f52-89ac-29702c46b57d",
      "type": "field_content_value",
      "created_at": "2018-06-07T08:48:21.836Z",
      "creator_user_id": "b.jankowski",
      "title": "Bledny login lub haso 1",
      "parameters": {
        "grace": 0,
        "backlog": 1,
        "repeat_notifications": true,
        "field": "action",
        "value": "Login incorrect (mschap: MS-CHAP2-Response is incorrect)"
      }
    },
    "triggered_at": "2018-06-07T10:52:11.232Z",
    "triggered": true,
    "matching_messages": [
      {
        "index": "eduroam_18",
        "message": "radius1 radiusd[51288]: (1548821)  Login incorrect (mschap: MS-CHAP2-Response is incorrect): [71072700875@uw.edu.pl] (from client trapeze-mx1 port 56454 cli 9C-F3-87-1E-DD-37 via TLS tunnel)",
        "fields": {
          "level": 5,
          "gl2_remote_ip": "10.30.87.42",
          "gl2_remote_port": 41690,
          "source-user": "71072700875@uw.edu.pl",
          "gl2_source_input": "55512e88e4b02f16ad5339c7",
          "EDUROAM_ACT": ": (1548821)  Login incorrect (mschap: MS-CHAP2-Response is incorrect): [71072700875@uw.edu.pl] (from client trapeze-mx1 port 56454 cli 9C-F3-87-1E-DD-37",
          "WINDOWSMAC": "9C-F3-87-1E-DD-37",
          "source-mac": "9C-F3-87-1E-DD-37",
          "Pesel": "71072700875",
          "Username": "71                                                                                                010.030.061.024.41245-010.012.003.236.00080: 072700875",
          "USERNAME": "trapeze-mx1",
          "action": "Login incorrect (mschap: MS-CHAP2-Response is incorrect)",
          "client": "trapeze-mx1",
          "gl2_source_node": "64f19870-4111-42dd-aef2-e7d662535efb",
          "facility": "local1",
          "Realm": "uw.edu.pl"
        },
        "id": "d1144d61-6a40-11e8-9f4a-00155d413d35",
        "timestamp": "2018-06-07T10:52:05.000Z",
        "source": "radius1",
        "stream_ids": [
          "5b02e93cce3109025090c1dd"
        ]
      }
    ]
  },
  "stream": {
    "creator_user_id": "p.zoltowski",
    "outputs": [],
    "description": "Logi radiusa eduroam",
    "created_at": "2018-05-21T15:43:56.012Z",
    "rules": [
      {
        "field": "source",
        "stream_id": "5b02e93cce3109025090c1dd",
        "description": "",
        "id": "5b02e9acce3109025090c25a",
        "type": 2,
        "inverted": false,
        "value": "radius[1-2]"
      }
    ],
    "alert_conditions": [
      {
        "creator_user_id": "b.jankowski",
        "created_at": "2018-06-07T08:48:21.836Z",
        "id": "9242a930-183a-4f52-89ac-29702c46b57d",
        "type": "field_content_value",
        "title": "Bledny login lub haso 1",
        "parameters": {
          "grace": 0,
          "backlog": 1,
          "repeat_notifications": true,
          "field": "action",
          "value": "Login incorrect (mschap: MS-CHAP2-Response is incorrect)"
        }
      }
    ],
    "title": "Eduroam",
    "content_pack": null,
    "is_default_stream": false,
    "index_set_id": "5b02e70fce3109025090bf77",
    "matching_type": "AND",
    "remove_matches_from_default_stream": true,
    "disabled": false,
    "id": "5b02e93cce3109025090c1dd"
  }
}
```

## Code Layout

The directory structure of the application:

    conf/             Configuration directory
        app.conf      Main app configuration file
        routes        Routes definition file

    app/              App sources
        init.go       Interceptor registration
        controllers/  App controllers go here
        views/        Templates directory

    messages/         Message files

    public/           Public static assets
        css/          CSS files
        js/           Javascript files
        images/       Image files

    tests/            Test suites
