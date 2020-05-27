package main

func (s *Server) routes() {
	s.router.HandleFunc("/register", s.handleregisteruser()).Methods("POST") // Unit Tested

}
