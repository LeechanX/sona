<!DOCTYPE html>
<html>
<head>
    <title>Sona</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
    <link rel="stylesheet" href="http://netdna.bootstrapcdn.com/bootstrap/3.0.3/css/bootstrap.min.css">
    <link rel="stylesheet" href="http://netdna.bootstrapcdn.com/bootstrap/3.0.3/css/bootstrap-theme.min.css">
    {{.HtmlHead}}
</head>
<body>

    <div class="container">
        <h3>Sona Configure Center <small>v0.0.1</small></h3>
        <br>
      </div>

    <form class="form-horizontal" action="/run/query" method="get" onsubmit="return check_submit()">
      <div class="form-group">
        <label for="service_key_id" class="col-sm-3 control-label">manage some serivce Key</label>
        <div class="col-sm-5">
          <input type="text" class="form-control" id="service_key_id" placeholder="productName.GroupName.ServiceName"
          name = "servicekey">
        </div>
        <div class="col-sm-1">
          <button type="submit" class="btn btn-default">Query</button>
        </div>
      </div>
      <div class="form-group">
        <label class="col-sm-3 control-label"><a href="/add">or add new serivce key </a></label>
      </div>
    </form>

    <script type="text/javascript" src="http://code.jquery.com/jquery-2.0.3.min.js"></script>
    <script src="http://netdna.bootstrapcdn.com/bootstrap/3.0.3/js/bootstrap.min.js"></script>
    <script type="text/javascript">
      function check_submit() {
        var key = $("#service_key_id").val()
        if (key == "") {
          alert("serivce key is empty")
          return false
        }
        var attrs = key.split(".")
        if (attrs.length != 3) {
          alert("serivce key's format error")
          return false
        }
        return true
      }
    </script>

</body>
</html>