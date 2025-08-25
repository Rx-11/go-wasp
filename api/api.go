package api

import (
	"fmt"
	"io/ioutil"

	"github.com/Rx-11/go-wasp/internal"
	"github.com/Rx-11/go-wasp/registry"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	reg registry.Registry
}

func NewServer(reg registry.Registry) *Server {
	return &Server{reg: reg}
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

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to read file")
	}

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
	wasmBytes, err := s.reg.GetFunction(name)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "function not found")
	}

	input := make(map[string]interface{})
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON input")
	}

	out, err := internal.ExecuteWASM(wasmBytes, input)
	if err != nil {
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
