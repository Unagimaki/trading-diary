import { apiRequest } from '@/shared/api/http'
export type ColumnType='text'|'number'|'select'|'date'|'boolean'|'image'
export type ColumnRole='trade_result'|'pnl'|'r'|null
export type JournalColumn={id:string;journalId:string;name:string;type:ColumnType;role:ColumnRole;options:string[];position:number;createdAt:string;updatedAt:string}
export type ColumnValues={name:string;type:ColumnType;role:ColumnRole;options:string[];position?:number}
export const columnsApi={list:(journalId:string)=>apiRequest<JournalColumn[]>(`/journals/${journalId}/columns`),create:(journalId:string,data:ColumnValues)=>apiRequest<JournalColumn>(`/journals/${journalId}/columns`,{method:'POST',body:JSON.stringify(data)}),update:(journalId:string,id:string,data:ColumnValues)=>apiRequest<JournalColumn>(`/journals/${journalId}/columns/${id}`,{method:'PATCH',body:JSON.stringify(data)}),remove:(journalId:string,id:string)=>apiRequest<void>(`/journals/${journalId}/columns/${id}`,{method:'DELETE'})}
