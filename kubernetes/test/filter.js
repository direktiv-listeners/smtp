if (event["type"] == "smtp.message") {
    nslog("rename type")
    event["type"] = "hello"
  }
  
return event