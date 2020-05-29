package main

func (s *Server) routes() {
	s.router.HandleFunc("/register", s.handleregisteruser()).Methods("POST")
	s.router.HandleFunc("/user", s.handleupdateuser()).Methods("PUT")
	s.router.HandleFunc("/login", s.handleloginuser()).Methods("POST")
	s.router.HandleFunc("/userpassword", s.handlechangeuserpassword()).Methods("PUT")
}
