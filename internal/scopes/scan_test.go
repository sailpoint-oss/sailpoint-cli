package scopes

import (
	"reflect"
	"testing"
)

func testManifest() *Manifest {
	return &Manifest{
		SchemaVersion: SchemaVersion,
		Operations: map[string]Operation{
			"listAccountsV1": {
				HTTPMethod: "GET", Path: "/accounts/v1",
				SDKNames: map[Language]string{
					LangGo: "ListAccountsV1", LangPython: "list_accounts_v1",
					LangTypeScript: "listAccountsV1", LangPowerShell: "Get-AccountsV1",
				},
				UserAuth:        []string{"idn:accounts:read", "idn:accounts:manage"},
				ApplicationAuth: []string{"idn:accounts:read", "idn:accounts:manage"},
			},
			"deleteAccountV1": {
				HTTPMethod: "DELETE", Path: "/accounts/v1/{id}",
				SDKNames: map[Language]string{
					LangGo: "DeleteAccountV1", LangPython: "delete_account_v1",
					LangTypeScript: "deleteAccountV1", LangPowerShell: "Remove-AccountV1",
				},
				UserAuth: []string{"idn:accounts:manage"},
			},
		},
	}
}

func TestScanFile(t *testing.T) {
	m := testManifest()
	tests := []struct {
		name string
		lang Language
		file sourceFile
		want []string
	}{
		{
			name: "go",
			lang: LangGo,
			file: sourceFile{Path: "main.go", Ext: ".go", Content: "import sailpoint \"github.com/sailpoint-oss/golang-sdk/v3\"\n" +
				"// apiClient.AccountsAPI.DeleteAccountV1(ctx, id)\n" +
				"resp, _, err := apiClient.AccountsAPI.ListAccountsV1(ctx).Limit(10).Execute()\n"},
			want: []string{"listAccountsV1"},
		},
		{
			name: "python with variants",
			lang: LangPython,
			file: sourceFile{Path: "app.py", Ext: ".py", Content: "from sailpoint import AccountsApi\n" +
				"# api.delete_account_v1(id)\n" +
				"api.list_accounts_v1()\n" +
				"api.delete_account_v1_with_http_info(id=x)\n"},
			want: []string{"listAccountsV1", "deleteAccountV1"},
		},
		{
			name: "typescript",
			lang: LangTypeScript,
			file: sourceFile{Path: "app.ts", Ext: ".ts", Content: "import { AccountsApi } from '@sailpoint/api-client';\n" +
				"/* api.deleteAccountV1({id}) */\n" +
				"const r = await api?.listAccountsV1({ limit: 1 });\n" +
				"other.listAccounts();\n"},
			want: []string{"listAccountsV1"},
		},
		{
			name: "powershell is case insensitive",
			lang: LangPowerShell,
			file: sourceFile{Path: "s.ps1", Ext: ".ps1", Content: "# Remove-AccountV1 -Id $x\n" +
				"$a = get-accountsv1 -Limit 1\n" +
				"Remove-AccountV1 -Id $a[0].id\n"},
			want: []string{"listAccountsV1", "deleteAccountV1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := scanFile(tt.file, tt.lang, m.NameIndex(tt.lang), m.Operations)
			var got []string
			for _, c := range calls {
				got = append(got, c.OperationID)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("operations = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileLanguage(t *testing.T) {
	used := usedLanguages([]sourceFile{
		{Ext: ".ts", Content: "import { AccountsService } from '@sailpoint/angular-sdk/accounts';"},
		{Ext: ".go", Content: "package main"},
	})
	if !used[LangAngular] || used[LangTypeScript] || used[LangGo] {
		t.Fatalf("used = %v, want only angular", used)
	}
	if lang, ok := fileLanguage(sourceFile{Ext: ".ts", Content: "this.svc.listAccountsV1()"}, used); !ok || lang != LangAngular {
		t.Errorf("ts file language = %q, %v; want angular", lang, ok)
	}
	if _, ok := fileLanguage(sourceFile{Ext: ".go", Content: "x.ListAccountsV1()"}, used); ok {
		t.Error("go file scanned although the project does not import the Go SDK")
	}
}
