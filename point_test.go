package redis

import (
	"context"
	"encoding/json"
	"github.com/go-tron/redis/script"
	"reflect"
	"testing"
)

var orderId int64 = 1502121461660254225

func TestPointGet(t *testing.T) {
	keys := []string{"point"}
	result, err := script.PointGet.Run(context.Background(), client, keys, 1).Result()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(reflect.TypeOf(result).Kind(), result)
}

func TestPointList(t *testing.T) {
	var data = []interface{}{
		1, 2, 3,
	}
	str, _ := json.Marshal(data)
	keys := []string{"point"}
	result, err := script.PointList.Run(context.Background(), client, keys, string(str)).Result()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(reflect.TypeOf(result).Kind(), result)
}

func TestPointBatchEdit(t *testing.T) {
	var data = [][]interface{}{
		{1, 10},
		{2, -20},
	}
	str, _ := json.Marshal(data)
	keys := []string{"point", "point-edit"}
	result, err := script.PointBatchEdit.Run(context.Background(), client, keys, orderId, string(str)).Result()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(reflect.TypeOf(result).Kind(), result)
}

func TestPointBatchRevoke(t *testing.T) {
	keys := []string{"point", "point-edit", "point-edit-revoke"}
	result, err := script.PointBatchRevoke.Run(context.Background(), client, keys, orderId).Result()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(reflect.TypeOf(result).Kind(), result)
}

func TestPointBatchApply(t *testing.T) {
	var data = [][]interface{}{
		{2, -20},
	}
	str, _ := json.Marshal(data)
	keys := []string{"point", "point-apply"}
	result, err := script.PointBatchApply.Run(context.Background(), client, keys, orderId, string(str)).Result()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(reflect.TypeOf(result).Kind(), result)
}

func TestPointBatchApplyConfirm(t *testing.T) {
	var data = [][]interface{}{}
	str, _ := json.Marshal(data)
	keys := []string{"point", "point-apply", "point-apply-result"}
	result, err := script.PointBatchApplyConfirm.Run(context.Background(), client, keys, orderId, string(str)).Result()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(reflect.TypeOf(result).Kind(), result)
}

func TestPointBatchApplyCancel(t *testing.T) {
	var data = [][]interface{}{}
	str, _ := json.Marshal(data)
	keys := []string{"point", "point-apply", "point-apply-result"}
	result, err := script.PointBatchApplyCancel.Run(context.Background(), client, keys, orderId, string(str)).Result()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(reflect.TypeOf(result).Kind(), result)
}

func TestPointBatchApplyRevoke(t *testing.T) {
	keys := []string{"point", "point-apply-result", "point-apply-revoke"}
	result, err := script.PointBatchRevoke.Run(context.Background(), client, keys, orderId).Result()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(reflect.TypeOf(result).Kind(), result)
}

func TestPointEdit(t *testing.T) {
	keys := []string{"point", "point-edit"}
	result, err := script.PointEdit.Run(context.Background(), client, keys, orderId, "0-1", 100).Result()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(reflect.TypeOf(result).Kind(), result)
}

func TestPointEditRevoke(t *testing.T) {
	keys := []string{"point", "point-edit", "point-edit-revoke"}
	result, err := script.PointRevoke.Run(context.Background(), client, keys, orderId).Result()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(reflect.TypeOf(result).Kind(), result)
}

func TestPointApply(t *testing.T) {
	keys := []string{"point", "point-apply"}
	var data = map[string]interface{}{
		"a": 1,
		"b": 2,
	}
	str, _ := json.Marshal(data)

	result, err := script.PointApply.Run(context.Background(), client, keys, orderId, "0-1", -10, str).Result()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(reflect.TypeOf(result).Kind(), result)
}

func TestPointApplyConfirm(t *testing.T) {
	keys := []string{"point", "point-apply", "point-apply-result", "point-edit"}
	result, err := script.PointApplyConfirm.Run(context.Background(), client, keys, orderId, "123458").Result()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(reflect.TypeOf(result).Kind(), result)
}

func TestPointApplyCancel(t *testing.T) {
	keys := []string{"point", "point-apply", "point-apply-result"}
	result, err := script.PointApplyCancel.Run(context.Background(), client, keys, orderId).Result()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(reflect.TypeOf(result).Kind(), result)
}

func TestPointApplyRevoke(t *testing.T) {
	keys := []string{"point", "point-apply-result", "point-apply-revoke"}
	result, err := script.PointRevoke.Run(context.Background(), client, keys, orderId).Result()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(reflect.TypeOf(result).Kind(), result)
}
