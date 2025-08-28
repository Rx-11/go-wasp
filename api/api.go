package api

import (
	"fmt"
	"io"

	"github.com/Rx-11/go-wasp/executor"
	"github.com/Rx-11/go-wasp/registry"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	reg        registry.Registry
	dispatcher *executor.Dispatcher
}

func NewServer(reg registry.Registry, disp *executor.Dispatcher) *Server {
	return &Server{reg: reg, dispatcher: disp}
}

func (s *Server) Routes() *fiber.App {
	app := fiber.New()
	app.Post("/upload", s.uploadFunction)
	app.Post("/:name/invoke", s.invokeFunction)
	app.Get("/functions", s.listFunctions)
	return app
}

func (s *Server) uploadFunction(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file required")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to open file")
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to read file")
	}

	fmt.Println(data[:100])

	name := c.FormValue("name")
	if name == "" {
		name = fileHeader.Filename
	}

	if err := s.reg.SaveFunction(name, data); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to save function")
	}

	return c.Status(fiber.StatusCreated).SendString(fmt.Sprintf("Function %s uploaded", name))
}

func (s *Server) invokeFunction(c *fiber.Ctx) error {
	name := c.Params("name")

	var input map[string]any
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&input); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid JSON input")
		}
	} else {
		input = map[string]any{}
	}

	out, err := s.dispatcher.Invoke(name, input)
	if err != nil {
		if err == executor.ErrQueueFull {
			return fiber.NewError(fiber.StatusTooManyRequests, "queue full for function")
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("execution error: %v", err))
	}

	return c.JSON(out)
}

func (s *Server) listFunctions(c *fiber.Ctx) error {
	funcs, err := s.reg.ListFunctions()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list functions")
	}
	return c.JSON(funcs)
}
