package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"

	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/internal/cli"
)

var (
	sessionServerAddr  string
	sessionJSON        bool
	sessionKeyword     string
	sessionStatus      string
	sessionType        string
	sessionAssistantID uint
	sessionOffset      int
	sessionLimit       int
)

type sessionDetailOutput struct {
	Session   *contract.Session                `json:"session,omitempty"`
	Assistant *contract.DigitalAssistantDetail `json:"assistant,omitempty"`
	Allocated *contract.DigitalAssistantDetail `json:"allocated_assistant,omitempty"`
	Messages  *contract.MessageList            `json:"messages,omitempty"`
	UserName  string                           `json:"user_name,omitempty"`
}

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage sessions",
	Long:  `Manage sessions in the Leros platform.`,
}

var sessionLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List sessions",
	Long:  `List all sessions with optional filtering.`,
	Run: func(cmd *cobra.Command, args []string) {
		go func() {
			req := &contract.ListSessionsRequest{
				Pagination: contract.ListSessionsRequest{}.Pagination,
			}
			req.Offset = sessionOffset
			req.Limit = sessionLimit
			req.Fill()

			if sessionKeyword != "" {
				req.Keyword = &sessionKeyword
			}
			if sessionStatus != "" {
				req.Status = &sessionStatus
			}
			if sessionType != "" {
				req.Type = &sessionType
			}
			if cmd.Flags().Changed("assistant-id") {
				req.AssistantID = &sessionAssistantID
			}

			result, err := cli.ListSessions(lifecycle.Std().Context(), sessionServerAddr, req)
			if err != nil {
				logs.Errorf("list sessions: %v", err)
				lifecycle.Std().Exit()
				return
			}
			printSessions(result)
			lifecycle.Std().Exit()
		}()
		lifecycle.Std().WaitExit()
	},
}

var sessionGetCmd = &cobra.Command{
	Use:   "get <session_id>",
	Short: "Get session details",
	Long:  `Get detailed information about a specific session by its public ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		go func() {
			ctx := lifecycle.Std().Context()
			sessionID := args[0]

			sess, err := cli.GetSession(ctx, sessionServerAddr, sessionID)
			if err != nil {
				logs.Errorf("get session: %v", err)
				lifecycle.Std().Exit()
				return
			}

			out := sessionDetailOutput{Session: sess}

			if sess.Uin > 0 {
				out.UserName = cli.ResolveUserName(ctx, sessionServerAddr, sess.Uin)
			}

			if sess.AssistantID > 0 {
				ast, err := cli.GetDigitalAssistantByID(ctx, sessionServerAddr, sess.AssistantID)
				if err != nil {
					logs.Warnf("get assistant: %v", err)
				} else {
					out.Assistant = ast
				}
			}
			if sess.AllocatedAssistantID > 0 && sess.AllocatedAssistantID != sess.AssistantID {
				ast, err := cli.GetDigitalAssistantByID(ctx, sessionServerAddr, sess.AllocatedAssistantID)
				if err != nil {
					logs.Warnf("get allocated assistant: %v", err)
				} else {
					out.Allocated = ast
				}
			}

			msgs, err := cli.GetSessionMessages(ctx, sessionServerAddr, sessionID, 1, 10)
			if err != nil {
				logs.Warnf("get session messages: %v", err)
			} else {
				out.Messages = msgs
			}

			printSessionDetail(&out)
			lifecycle.Std().Exit()
		}()
		lifecycle.Std().WaitExit()
	},
}

func printSessions(list *contract.SessionList) {
	if sessionJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(list.Items)
		return
	}

	if len(list.Items) == 0 {
		fmt.Println("No sessions found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "PUBLIC_ID\tTYPE\tSTATUS\tTITLE\tMSG_COUNT\tCREATED_AT")
	for _, s := range list.Items {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
			s.SessionID, s.Type, s.Status, s.Title, s.MessageCount,
			s.CreatedAt.Format("2006-01-02T15:04:05Z"))
	}
	w.Flush()

	fmt.Fprintf(os.Stderr, "\nTotal: %d, Offset: %d, Limit: %d\n", list.Total, list.Offset, list.Limit)
}

func printSessionDetail(out *sessionDetailOutput) {
	if sessionJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(out)
		return
	}

	s := out.Session
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "SessionID:\t%s\n", s.SessionID)
	fmt.Fprintf(w, "Type:\t%s\n", s.Type)
	fmt.Fprintf(w, "Status:\t%s\n", s.Status)
	fmt.Fprintf(w, "Title:\t%s\n", s.Title)
	fmt.Fprintf(w, "TitleManuallySet:\t%v\n", s.TitleManuallySet)
	if out.UserName != "" {
		fmt.Fprintf(w, "User:\t%s (uin=%d)\n", out.UserName, s.Uin)
	} else {
		fmt.Fprintf(w, "Uin:\t%d\n", s.Uin)
	}
	fmt.Fprintf(w, "OrgID:\t%d\n", s.OrgID)

	if out.Assistant != nil {
		fmt.Fprintf(w, "Assistant:\t%s (ID=%d, Code=%s)\n", out.Assistant.Name, out.Assistant.ID, out.Assistant.Code)
	} else {
		fmt.Fprintf(w, "AssistantID:\t%d\n", s.AssistantID)
	}

	if out.Allocated != nil {
		fmt.Fprintf(w, "AllocatedAssistant:\t%s (ID=%d, Code=%s)\n", out.Allocated.Name, out.Allocated.ID, out.Allocated.Code)
	} else if s.AllocatedAssistantID > 0 {
		fmt.Fprintf(w, "AllocatedAssistantID:\t%d\n", s.AllocatedAssistantID)
	}

	fmt.Fprintf(w, "AssistantCode:\t%s\n", s.AssistantCode)
	fmt.Fprintf(w, "MessageCount:\t%d\n", s.MessageCount)
	if s.LastMessageAt != nil {
		fmt.Fprintf(w, "LastMessageAt:\t%s\n", s.LastMessageAt.Format("2006-01-02T15:04:05Z"))
	}
	if s.ExpiredAt != nil {
		fmt.Fprintf(w, "ExpiredAt:\t%s\n", s.ExpiredAt.Format("2006-01-02T15:04:05Z"))
	}
	fmt.Fprintf(w, "CreatedAt:\t%s\n", s.CreatedAt.Format("2006-01-02T15:04:05Z"))
	fmt.Fprintf(w, "UpdatedAt:\t%s\n", s.UpdatedAt.Format("2006-01-02T15:04:05Z"))

	if out.Messages != nil && len(out.Messages.Items) > 0 {
		fmt.Fprintf(w, "--- Messages (last %d/%d) ---\t\n", len(out.Messages.Items), out.Messages.Total)
		for i, m := range out.Messages.Items {
			content := truncateContent(m.Content, 120)
			fmt.Fprintf(w, "  [%d] %s\t%s\n", i+1, m.Role, content)
		}
	}

	w.Flush()
}

func truncateContent(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func init() {
	sessionCmd.PersistentFlags().StringVar(&sessionServerAddr, "server-addr", "127.0.0.1:8080", "Leros server address (host:port)")
	sessionCmd.PersistentFlags().BoolVar(&sessionJSON, "json", false, "Output in JSON format")

	sessionLsCmd.Flags().StringVar(&sessionKeyword, "keyword", "", "Filter by title or public_id keyword")
	sessionLsCmd.Flags().StringVar(&sessionStatus, "status", "", "Filter by status")
	sessionLsCmd.Flags().StringVar(&sessionType, "type", "", "Filter by session type")
	sessionLsCmd.Flags().UintVar(&sessionAssistantID, "assistant-id", 0, "Filter by assistant ID")
	sessionLsCmd.Flags().IntVar(&sessionOffset, "offset", 0, "Pagination offset")
	sessionLsCmd.Flags().IntVar(&sessionLimit, "limit", 20, "Pagination limit")

	sessionCmd.AddCommand(sessionLsCmd)
	sessionCmd.AddCommand(sessionGetCmd)
	rootCmd.AddCommand(sessionCmd)
}
