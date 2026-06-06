package consistentreceivers

type Handler struct{}

func (h *Handler) ServeHTTP() {}

func (handler *Handler) Handle() {} // want "receiver name .* inconsistent"

func (h *Handler) Process() {}

type Server struct{}

func (s *Server) Start() {}

func (s *Server) Stop() {}
