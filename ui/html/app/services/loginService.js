app
.service("loginService",["requestService", "$http",function(requestService, $http){
    this.login = function(username,password){
        requestService.post_request("/api/user/login",{
            username: username,
            password: password}, "").then(function(response){
                if(response.data.code == 200){
                    setSession("token",response.data.session_token);
                    if(response.data.hasOwnProperty("is_admin")){
                        setSession("is_admin",true);
                        window.location.href = "admin.html#!/get_services";
                    }else{
                        setSession("is_admin",false);
                        window.location.href = "/dashboard.html#!/all_jobs";
                    }
                }else{
                    alert(response.data.description);
                }
            });
    }
    // save  session token
    function setSession (name, value){
        window.sessionStorage.setItem(name,value);
    }
    return this;
}])