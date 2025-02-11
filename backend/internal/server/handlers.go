package server

import (
	"context"
	"fmt"
	"time"

	"github.com/kxrxh/ddss/internal/db"
)

func (s *Server) handleRequest(req Request) Response {
	switch req.Action {
	case "getCitizen":
		return s.handleGetCitizen(req.Data)
	case "updateScore":
		return s.handleUpdateScore(req.Data)
	case "addEvent":
		return s.handleAddEvent(req.Data)
	default:
		return Response{
			Success: false,
			Error:   "Unknown action",
		}
	}
}

func (s *Server) handleGetCitizen(data map[string]any) Response {
	id, ok := data["id"].(string)
	if !ok {
		return Response{
			Success: false,
			Error:   "Invalid citizen ID format",
		}
	}

	citizen, err := s.dgraph.GetCitizen(context.Background(), id)
	if err != nil {
		return Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to get citizen: %v", err),
		}
	}

	return Response{
		Success: true,
		Data:    citizen,
	}
}

func (s *Server) handleUpdateScore(data map[string]any) Response {
	id, ok := data["id"].(string)
	if !ok {
		return Response{
			Success: false,
			Error:   "Invalid citizen ID format",
		}
	}

	score, ok := data["score"].(float64)
	if !ok {
		return Response{
			Success: false,
			Error:   "Invalid score format",
		}
	}

	err := s.dgraph.UpdateScore(context.Background(), id, score)
	if err != nil {
		return Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to update score: %v", err),
		}
	}

	return Response{
		Success: true,
	}
}

func (s *Server) handleAddEvent(data map[string]any) Response {
	citizenID, ok := data["citizen_id"].(string)
	if !ok {
		return Response{
			Success: false,
			Error:   "Invalid citizen ID format",
		}
	}

	eventType, ok := data["type"].(string)
	if !ok {
		return Response{
			Success: false,
			Error:   "Invalid event type",
		}
	}

	scoreChange, ok := data["score_change"].(float64)
	if !ok {
		return Response{
			Success: false,
			Error:   "Invalid score change value",
		}
	}

	description, _ := data["description"].(string)

	event := db.Event{
		Type:        eventType,
		ScoreChange: scoreChange,
		Description: description,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	err := s.dgraph.AddEvent(context.Background(), citizenID, event)
	if err != nil {
		return Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to add event: %v", err),
		}
	}

	return Response{
		Success: true,
		Data:    event,
	}
}
