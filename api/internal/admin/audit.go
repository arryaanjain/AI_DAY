package admin

import("context";"github.com/jackc/pgx/v5/pgxpool")
type AuditLog struct{ActorUserID,Action,ResourceType,ResourceID string;Metadata []byte}
type Service struct{db *pgxpool.Pool}
func New(db *pgxpool.Pool)*Service{return &Service{db:db}}
func(s *Service)Record(ctx context.Context,entry AuditLog)error{_,err:=s.db.Exec(ctx,`INSERT INTO audit_logs(actor_user_id,action,resource_type,resource_id,metadata) VALUES(NULLIF($1,'')::uuid,$2,$3,NULLIF($4,'')::uuid,COALESCE($5::jsonb,'{}'::jsonb))`,entry.ActorUserID,entry.Action,entry.ResourceType,entry.ResourceID,string(entry.Metadata));return err}
