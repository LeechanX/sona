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

    <form id="add_form" class="form-horizontal" action="/run/add" method="post" onsubmit="return check_submit()">
      <div class="form-group">
        <label for="inputEmail3" class="col-sm-2 control-label">product name </label>
        <div class="col-sm-3">
          <input type="text" class="form-control" id="product_id" name="product_name">
        </div>
      </div>
      <div class="form-group">
        <label for="inputPassword3" class="col-sm-2 control-label">group name </label>
        <div class="col-sm-3">
          <input type="text" class="form-control" id="group_id" name="group_name">
        </div>
      </div>
      <div class="form-group">
        <label for="inputPassword3" class="col-sm-2 control-label">service name </label>
        <div class="col-sm-3">
          <input type="text" class="form-control" id="service_id" name="service_name">
        </div>
      </div>
      <div class="form-group">
        <div class="col-sm-offset-2 col-sm-10">
          <button type="submit" class="btn btn-info">Add Service Key</button>
        </div>
      </div>
      <blockquote class="col-sm-offset-2 col-sm-10" id="result">
      </blockquote>
    </form>

    <script type="text/javascript" src="http://code.jquery.com/jquery-2.0.3.min.js"></script>
    <script src="http://netdna.bootstrapcdn.com/bootstrap/3.0.3/js/bootstrap.min.js"></script>
    <script src="http://malsup.github.com/jquery.form.js"></script> 

    <script type="text/javascript">
      function check_submit() {
        var reg = /^\w+$/;
        var key = $("#product_id").val()
        if (key == "") {
          alert("product name is empty")
          return false
        }
        if (key.length > 30) {
          alert("product name length exceed 30 bytes")
          return false
        }
        if (reg.test(key) == false) {
          alert("product name format error")
          return false
        }
        key = $("#group_id").val()
        if (key == "") {
          alert("group name is empty")
          return false
        }
        if (key.length > 30) {
          alert("group name length exceed 30 bytes")
          return false
        }
        if (reg.test(key) == false) {
          alert("group name format error")
          return false
        }
        key = $("#service_id").val()
        if (key == "") {
          alert("serivce name is empty")
          return false
        }
        if (key.length > 30) {
          alert("serivce name is empty")
          return false
        }
        if (reg.test(key) == false) {
          alert("serivce name format error")
          return false
        }
        return true
      }

      $("#add_form").ajaxForm(function(data) {
        msg = eval(data)
        if(msg.code == 200) {
          $("#result").html("<p class=\"text-success\">created successfully</p>")
        } else {
          $("#result").html("<p class=\"text-danger\">created error: " + msg.message + "</p>")
        }
      });

    </script>

</body>
</html>