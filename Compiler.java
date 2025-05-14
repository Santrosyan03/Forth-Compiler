import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.*;

public class Compiler {

    private Deque<Object> stack = new ArrayDeque<>();
    private Map<String, Integer> variables = new HashMap<>();

    public void execute(String code) {
        String[] lines = code.split("\\r?\\n");
        for (String line : lines) {
            line = line.split("\\\\")[0].trim();
            if (line.isEmpty()) continue;

            String[] tokens = line.split("\\s+");
            for (int i = 0; i < tokens.length; i++) {
                String token = tokens[i];

                switch (token) {
                    case "variable":
                        if (i + 1 >= tokens.length) {
                            System.err.println("Error: Missing variable identifier after 'variable'");
                            break;
                        }
                        i++;
                        push(tokens[i]);
                        variable();
                        break;

                    case "!":
                    case "store":
                        store();
                        break;

                    case "@":
                    case "fetch":
                        fetch();
                        break;

                    case "+":
                        add();
                        break;

                    case "-":
                        subtract();
                        break;

                    case "*":
                        multiply();
                        break;

                    case "dup":
                        dup();
                        break;

                    case "swap":
                        swap();
                        break;

                    case "drop":
                        drop();
                        break;

                    case "over":
                        over();
                        break;

                    case "mod":
                        mod();
                        break;

                    case "neg":
                        neg();
                        break;

                    case "nip":
                        nip();
                        break;

                    case "tuck":
                        tuck();
                        break;

                    case ".":
                        printTop();
                        break;

                    case ".s":
                        printStack();
                        break;

                    default:
                        try {
                            int val = Integer.parseInt(token);
                            push(val);
                        } catch (NumberFormatException e) {
                            push(token);
                        }
                        break;
                }
            }
        }
    }

    private void push(Object val) {
        stack.push(val);
    }

    private void variable() {
        if (stack.isEmpty()) return;
        Object varName = stack.pop();
        if (!(varName instanceof String)) {
            System.err.println("Invalid var name: " + varName + ". Expected a string.");
            return;
        }
        variables.putIfAbsent((String) varName, 0);
    }

    private void store() {
        if (stack.size() < 2) {
            System.err.println("Error: Insufficient stack elements for 'store' operation.");
            return;
        }
        Object varName = stack.pop();
        Object value = stack.pop();
        if (!(varName instanceof String)) {
            System.err.println("Invalid var name: " + varName + ". Must be a string.");
            return;
        }
        if (!variables.containsKey(varName)) {
            System.err.println("Unknown variable '" + varName + "'. Declare it before storing.");
            return;
        }
        if (!(value instanceof Integer)) {
            System.err.println("Invalid value type: " + value + ". Integer required.");
            return;
        }
        variables.put((String) varName, (Integer) value);
    }

    private void fetch() {
        if (stack.isEmpty()) return;
        Object varName = stack.pop();
        if (!(varName instanceof String)) {
            System.err.println("Invalid var name: " + varName + ". Expected string.");
            return;
        }
        if (!variables.containsKey(varName)) {
            System.err.println("Variable '" + varName + "' not found. Declare it first.");
            return;
        }
        push(variables.get(varName));
    }

    private void add() {
        if (stack.size() < 2) return;
        int b = popInt();
        int a = popInt();
        push(a + b);
    }

    private void subtract() {
        if (stack.size() < 2) return;
        int b = popInt();
        int a = popInt();
        push(a - b);
    }

    private void multiply() {
        if (stack.size() < 2) return;
        int b = popInt();
        int a = popInt();
        push(a * b);
    }

    private void mod() {
        if (stack.size() < 2) return;
        int b = popInt();
        int a = popInt();
        if (b == 0) push(0);
        else push(a % b);
    }

    private void dup() {
        if (stack.isEmpty()) return;
        push(stack.peek());
    }

    private void swap() {
        if (stack.size() < 2) return;
        int a = popInt();
        int b = popInt();
        push(a);
        push(b);
    }

    private void drop() {
        if (!stack.isEmpty()) stack.pop();
    }

    private void over() {
        if (stack.size() < 2) return;
        Iterator<Object> it = stack.iterator();
        it.next();
        Object second = it.next();
        push(second);
    }

    private void neg() {
        if (stack.isEmpty()) return;
        int a = popInt();
        push(-a);
    }

    private void nip() {
        if (stack.size() < 2) return;
        int top = popInt();
        int second = popInt();
        push(top);
    }

    private void tuck() {
        if (stack.size() < 2) return;
        int top = popInt();
        int second = popInt();
        push(top);
        push(second);
        push(top);
    }

    private void printTop() {
        if (stack.isEmpty()) return;
        System.out.println(stack.pop());
    }

    private void printStack() {
        List<Object> list = new ArrayList<>(stack);
        Collections.reverse(list);
        System.out.println("[Bottom -> Top] " + list);
    }

    private int popInt() {
        Object val = stack.pop();
        if (val instanceof Integer) return (Integer) val;
        throw new RuntimeException("Stack error: Expected Integer but found " + val);
    }

    public static void main(String[] args) throws IOException {
        if (args.length < 1) {
            System.err.println("Provide a file name as argument.");
            return;
        }
        String str = new String(Files.readAllBytes(Paths.get(args[0])));
        Compiler c = new Compiler();
        c.execute(str);
    }
}
